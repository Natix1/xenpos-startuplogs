package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type StartupLog struct {
	IsStudio      bool `json:"is_studio"`
	FirstPlayerId int  `json:"first_player_id"`
}

var (
	DISCORD_WEBHOOK string
	REDIS           *redis.Client
	HTTP_CLIENT     *http.Client
)

func init() {
	godotenv.Load()

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		panic("REDIS_ADDR not specified")
	}

	DISCORD_WEBHOOK = os.Getenv("DISCORD_WEBHOOK")
	if DISCORD_WEBHOOK == "" {
		panic("DISCORD_WEBHOOK not specified")
	}

	client := redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	HTTP_CLIENT = &http.Client{}
	REDIS = client

	if REDIS.Ping(context.Background()).Err() != nil {
		panic("Couldn't connect to redis")
	}
}

func httpError(w http.ResponseWriter, status int, err string) {
	http.Error(w, http.StatusText(status)+": "+err, status)
	slog.Error(err, "status code", status, "status", http.StatusText(status))
}

func makeRequest(method string, url string, body []byte) (*http.Response, error) {
	request, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	request.Header.Set("Content-Type", "application/json")
	resp, err := HTTP_CLIENT.Do(request)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		return nil, errors.New("Non-200 status code: " + strconv.Itoa(resp.StatusCode) + " " + string(body))
	}

	return resp, nil
}

// TODO https://create.roblox.com/docs/cloud/reference/Universe

func getUniverseId(placeId int) (int, error) {
	url := fmt.Sprintf("https://apis.roblox.com/universes/v1/places/%d/universe", placeId)
	resp, err := makeRequest("GET", url, []byte(""))
	if err != nil {
		return 0, err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var response struct {
		UniverseId int `json:"universeId"`
	}
	err = json.Unmarshal(body, &response)
	if err != nil {
		return 0, err
	}

	return response.UniverseId, nil
}

func makeRedisKey(key string) string {
	return "xenpos:startuplogs:" + key
}

func main() {
	http.HandleFunc("/pos/startup-log", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpError(w, http.StatusMethodNotAllowed, "POST only endpoint")
			return
		}

		robloxPlace := r.Header.Get("Roblox-Id")
		if robloxPlace == "" {
			httpError(w, http.StatusBadRequest, "Only requests from HttpService allowed")
			return
		}

		robloxPlaceId, err := strconv.Atoi(robloxPlace)
		if err != nil {
			httpError(w, http.StatusInternalServerError, "Failed parsing place id to base10 int")
			return
		}

		redisKey := makeRedisKey("placeid:" + robloxPlace)
		exists, err := REDIS.Exists(context.Background(), redisKey).Result()
		if err != nil {
			httpError(w, http.StatusInternalServerError, "Failed reaching database")
			return
		}

		if exists == 1 {
			w.Write([]byte("Already registered before"))
			slog.Info("already registered before", "place id", robloxPlace)
		} else {
			body, err := io.ReadAll(r.Body)
			if err != nil {
				httpError(w, http.StatusInternalServerError, "Failed opening body")
				return
			}

			defer r.Body.Close()
			var request StartupLog
			err = json.Unmarshal(body, &request)

			if err != nil {
				httpError(w, http.StatusBadRequest, "Failed parsing body")
				return
			}

			universeId, err := getUniverseId(robloxPlaceId)
			if err != nil {
				httpError(w, http.StatusBadRequest, "Invalid place id or something went wrong: "+err.Error())
				return
			}

			payload := map[string]any{
				"content": "@everyone",
				"embeds": []map[string]any{
					{
						"title":     "New game!",
						"color":     15277667,
						"timestamp": time.Now().Format(time.RFC3339),
						"fields": []map[string]any{
							{"name": "Is studio", "value": strconv.FormatBool(request.IsStudio), "inline": true},
							{"name": "First player id", "value": strconv.Itoa(request.FirstPlayerId), "inline": true},
							{"name": "Place id", "value": strconv.Itoa(robloxPlaceId), "inline": true},
							{"name": "Universe id", "value": strconv.Itoa(universeId), "inline": true},
						},
					},
				},
			}

			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				slog.Error(err.Error())
				httpError(w, http.StatusInternalServerError, "failed marshaling payload")
				return
			}

			resp, err := makeRequest("POST", DISCORD_WEBHOOK, jsonPayload)
			if err != nil {
				slog.Error(err.Error())
				httpError(w, http.StatusInternalServerError, "failed logging on the backend")
				return
			}
			defer resp.Body.Close()

			slog.Info("Recorded game",
				"studio", request.IsStudio,
				"first player id", request.FirstPlayerId,
				"place id", robloxPlace,
				"universe id", universeId,
			)

			if err := REDIS.Set(context.Background(), redisKey, 1, time.Hour*24*365).Err(); err != nil {
				slog.Error("failed to write redis key", "err", err, "key", redisKey)
				httpError(w, http.StatusInternalServerError, "failed to write redis")
				return
			}
			w.Write([]byte("Recorded"))
		}
	})

	http.ListenAndServe(":8080", nil)
}
