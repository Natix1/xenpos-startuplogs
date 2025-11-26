package main

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type StartupLog struct {
	IsStudio            bool `json:"is_studio"`
	StudioFirstPlayerId int  `json:"studio_first_player_id"`
	UniverseId          int  `json:"universe_id"`
	PlaceId             int  `json:"place_id"`
}

var (
	HOSTNAME        string
	DISCORD_WEBHOOK string
)

func init() {
	godotenv.Load()

	HOSTNAME = os.Getenv("HOSTNAME")
	if HOSTNAME == "" {
		panic("HOSTNAME not defined")
	}

	DISCORD_WEBHOOK = os.Getenv("DISCORD_WEBHOOK")
	if DISCORD_WEBHOOK == "" {

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

		w.Write([]byte("Recorded"))
	})

	http.ListenAndServe(":8080", nil)
}
