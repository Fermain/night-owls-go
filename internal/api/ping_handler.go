package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

// PingRequest is the expected JSON for POST /api/ping.
type PingRequest struct {
	Message string `json:"message"`
}

// PingResponse is the JSON response for POST /api/ping.
type PingResponse struct {
	Echo      string    `json:"echo"`
	Timestamp time.Time `json:"timestamp"`
}

// PingHandler handles the /api/ping endpoint.
func PingHandler(logger *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			RespondWithError(w, http.StatusMethodNotAllowed, "Only POST method is allowed", logger)
			return
		}

		var req PingRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			RespondWithError(w, http.StatusBadRequest, "Invalid request payload", logger, "error", err.Error())
			return
		}

		if req.Message == "" {
			RespondWithError(w, http.StatusBadRequest, "Message cannot be empty", logger)
			return
		}

		resp := PingResponse{
			Echo:      req.Message,
			Timestamp: time.Now(),
		}

		RespondWithJSON(w, http.StatusOK, resp, logger.With("handler", "PingHandler"))
	}
}
