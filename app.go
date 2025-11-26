package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	discordwebhook "github.com/bensch777/discord-webhook-golang"
	"github.com/joho/godotenv"
)

type StartupLog struct {
	IsStudio      bool `json:"is_studio"`
	FirstPlayerId int  `json:"first_player_id"`
	UniverseId    int  `json:"universe_id"`
	PlaceId       int  `json:"place_id"`
	CreatorId     int  `json:"creator_id"`
	CreatorType   int  `json:"creator_type"`
}

var (
	DISCORD_WEBHOOK string
)

func init() {
	godotenv.Load()

	DISCORD_WEBHOOK = os.Getenv("DISCORD_WEBHOOK")
	if DISCORD_WEBHOOK == "" {
		panic("DISCORD_WEBHOOK not specified")
	}
}

func httpError(w http.ResponseWriter, status int, err string) {
	http.Error(w, http.StatusText(status)+": "+err, status)
	slog.Error(err, "status code", status, "status", http.StatusText(status))
}

func main() {
	http.HandleFunc("/pos/startup-log", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			httpError(w, http.StatusMethodNotAllowed, "POST only endpoint")
			return
		}

		userAgentParts := strings.Split(r.UserAgent(), " ")
		if len(userAgentParts) < 1 {
			httpError(w, http.StatusBadRequest, "Only requests from HttpService allowed (mismatch no. 1)")
			return
		}

		robloxPlaceId := r.Header.Get("Roblox-Id")
		if robloxPlaceId == "" {
			httpError(w, http.StatusBadRequest, "Only requests from HttpService allowed (mismatch no. 2)")
			return
		}

		if !strings.EqualFold(userAgentParts[0], "RobloxGameCloud/1.0") {
			httpError(w, http.StatusBadRequest, "Invalid useragent supplied")
			return
		}

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

		if robloxPlaceId != strconv.Itoa(request.PlaceId) {
			httpError(w, http.StatusBadRequest, "Suspicious request (mismatch no. 3)")
			return
		}

		slog.Info("Recorded game",
			"studio", request.IsStudio,
			"first player id", request.FirstPlayerId,
			"universe id", request.UniverseId,
			"place id", request.PlaceId,
			"creator id", request.CreatorId,
			"creator type", request.CreatorType,
		)

		embed := discordwebhook.Embed{
			Title:     "New game!",
			Color:     15277667,
			Timestamp: time.Now(),
			Fields: []discordwebhook.Field{
				{
					Name:   "Is studio",
					Value:  strconv.FormatBool(request.IsStudio),
					Inline: true,
				},
				{
					Name:   "First player id",
					Value:  strconv.Itoa(request.FirstPlayerId),
					Inline: true,
				},
				{
					Name:   "Universe id",
					Value:  strconv.Itoa(request.UniverseId),
					Inline: true,
				},
				{
					Name:   "Place id",
					Value:  strconv.Itoa(request.PlaceId),
					Inline: true,
				},
				{
					Name:   "Creator id",
					Value:  strconv.Itoa(request.CreatorId),
					Inline: true,
				},
				{
					Name:   "Creator type",
					Value:  strconv.Itoa(request.CreatorType),
					Inline: true,
				},
			},
		}

		discordwebhook.SendEmbed(DISCORD_WEBHOOK, embed)
		w.Write([]byte("Recorded"))
	})

	http.ListenAndServe(":8080", nil)
}
