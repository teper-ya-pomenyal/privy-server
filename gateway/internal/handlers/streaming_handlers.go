package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/clients"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/middlewares"
)

type StreamHandlers struct {
	grpcClient *clients.UserClient
}

func NewStreamingHandlers(uc *clients.UserClient) *StreamHandlers {
	return &StreamHandlers{grpcClient: uc}
}

func (h *StreamHandlers) CreateStreamToken(w http.ResponseWriter, r *http.Request) {
	stringUUID := r.Context().Value(middlewares.UserIDKey).(string)
	userUUID, err := uuid.Parse(stringUUID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	birthDate := r.Context().Value(middlewares.BDKey).(time.Time)

	token, err := h.grpcClient.CreateStreamToken(r.Context(), userUUID, birthDate)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"stream_token": token})
}
