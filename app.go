package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
	"github.com/natix1/xenpos-startuplogs/src/rest"
	"github.com/natix1/xenpos-startuplogs/src/roblox"
	"github.com/redis/go-redis/v9"
)

type StartupLogRequest struct {
	IsStudio      bool `json:"is_studio"`
	FirstPlayerId int  `json:"first_player_id"`
}

var (
	DISCORD_WEBHOOK string
	REDIS_CLIENT    *redis.Client
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

	rest.ROBLOX_API_KEY = os.Getenv("OPEN_CLOUD_KEY")
	if rest.ROBLOX_API_KEY == "" {
		panic("OPEN_CLOUD_KEY not specified")
	}

	REDIS_CLIENT = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: "",
		DB:       0,
	})

	if REDIS_CLIENT.Ping(context.Background()).Err() != nil {
		panic("Couldn't connect to redis")
	}
}

func makeRedisKey(key string) string {
	return "xenpos:startuplogs:" + key
}

func httpError(w http.ResponseWriter, status int, err string) {
	http.Error(w, http.StatusText(status)+": "+err, status)
	slog.Error(err, "status code", status, "status", http.StatusText(status))
}

func startupLogHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		httpError(w, http.StatusMethodNotAllowed, "POST only endpoint")
		return
	}

	robloxPlaceStr := r.Header.Get("Roblox-Id")
	if robloxPlaceStr == "" {
		httpError(w, http.StatusBadRequest, "Only requests from HttpService allowed")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed opening body")
		return
	}

	defer r.Body.Close()
	var request StartupLogRequest
	err = json.Unmarshal(body, &request)

	if err != nil {
		httpError(w, http.StatusBadRequest, "Failed parsing body")
		return
	}

	robloxPlaceId, err := strconv.Atoi(robloxPlaceStr)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed parsing place id to base10 int")
		return
	}

	isStudioStr := strconv.FormatBool(request.IsStudio)
	redisKey := makeRedisKey(fmt.Sprintf("placeId=%d:isStudio=%s", robloxPlaceId, isStudioStr))
	exists, err := REDIS_CLIENT.Exists(context.Background(), redisKey).Result()
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed reaching database")
		return
	}

	if exists == 1 {
		w.Write([]byte("Already registered before"))
		slog.Info("already registered before", "place id", robloxPlaceStr)
		return
	}

	universeId, err := roblox.GetUniverseIdFromPlaceId(robloxPlaceId)
	if err != nil {
		httpError(w, http.StatusBadRequest, "Invalid place id or something went wrong: "+err.Error())
		return
	}

	universe, err := roblox.GetUniverse(universeId)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed getting universe data: "+err.Error())
		return
	}

	user, err := roblox.GetUser(request.FirstPlayerId)
	if err != nil {
		httpError(w, http.StatusInternalServerError, "Failed getting user data"+err.Error())
		return
	}

	content := fmt.Sprintf("# New place ID used in universe '%s' \n-# @everyone \nIn studio: **%s** \nFirst player: **%s (%s)** \nPlace ID: **%d** \nUniverse ID: **%d** \n-# %s",
		universe.DisplayName,
		isStudioStr,
		user.Name,
		user.ID,
		robloxPlaceId,
		universeId,
		time.Now().Format(time.RFC3339),
	)
	payload := map[string]any{
		"content": content,
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		slog.Error(err.Error())
		httpError(w, http.StatusInternalServerError, "failed marshaling payload")
		return
	}

	resp, err := http.Post(DISCORD_WEBHOOK, "application/json", bytes.NewBuffer(jsonPayload))
	if err != nil {
		slog.Error(err.Error())
		httpError(w, http.StatusInternalServerError, "failed logging on the backend")
		return
	}
	defer resp.Body.Close()

	slog.Info("Recorded game",
		"studio", request.IsStudio,
		"first player id", request.FirstPlayerId,
		"place id", robloxPlaceStr,
		"universe id", universeId,
		"universe name", universe.DisplayName,
	)

	if err := REDIS_CLIENT.Set(context.Background(), redisKey, 1, time.Hour*24*365).Err(); err != nil {
		slog.Error("failed to write redis key", "err", err, "key", redisKey)
		httpError(w, http.StatusInternalServerError, "failed to write redis")
		return
	}
	w.Write([]byte("Recorded"))
}

func main() {
	http.HandleFunc("/pos/startup-log", startupLogHandler)
	http.ListenAndServe(":8080", nil)
}
