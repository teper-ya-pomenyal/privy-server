package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/clients"
	mw "github.com/teper-ya-pomenyal/privy_stream/gateway/internal/middlewares"
)

// AdminHandler — закрытые ролью owner операции над узлом: сессии,
// пользователи, метки 18+, health и журнал. Бизнес-логика остаётся
// в сервисах: gateway только прокидывает вызовы по gRPC.
type AdminHandler struct {
	userClient    *clients.UserClient
	catalogClient *clients.CatalogClient
	logs          *AdminLogRing
	streamingAddr string
	storagePath   string
}

func NewAdminHandler(userClient *clients.UserClient, catalogClient *clients.CatalogClient, logs *AdminLogRing, streamingAddr, storagePath string) *AdminHandler {
	return &AdminHandler{
		userClient:    userClient,
		catalogClient: catalogClient,
		logs:          logs,
		streamingAddr: streamingAddr,
		storagePath:   storagePath,
	}
}

// MountRoutes монтируется в cmd/main.go рядом с /catalog и /stream.
func (h *AdminHandler) MountRoutes(r chi.Router, m *mw.MiddleWares) {
	r.Route("/admin", func(r chi.Router) {
		r.Use(m.Auth.Handle)
		r.Use(m.Owner.Handle)

		r.Get("/health", h.Health)
		r.Get("/logs", h.Logs)

		r.Get("/sessions", h.ListSessions)
		r.Delete("/sessions/{session_id}", h.DeleteSession)
		r.Post("/sessions/revoke-others", h.RevokeOthers)

		r.Get("/users", h.ListUsers)
		r.Patch("/users/{user_uuid}", h.PatchUser)

		r.Get("/moderation", h.ModerationList)
		r.Patch("/moderation/{track_uuid}", h.ModerationSetExplicit)
	})
}

func parseUUIDParam(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(chi.URLParam(r, key))
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return uuid.UUID{}, false
	}
	return id, true
}

func parseQueryUUID(w http.ResponseWriter, r *http.Request, key string) (uuid.UUID, bool) {
	id, err := uuid.Parse(r.URL.Query().Get(key))
	if err != nil {
		http.Error(w, "invalid "+key+" query parameter", http.StatusBadRequest)
		return uuid.UUID{}, false
	}
	return id, true
}

func decodeBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return false
	}
	return true
}

// ---------- sessions ----------

func (h *AdminHandler) ListSessions(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseQueryUUID(w, r, "user")
	if !ok {
		return
	}
	sessions, err := h.userClient.ListSessions(r.Context(), userUUID.String())
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"sessions": sessions})
}

func (h *AdminHandler) DeleteSession(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseQueryUUID(w, r, "user")
	if !ok {
		return
	}
	sessionID := chi.URLParam(r, "session_id")
	if sessionID == "" {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	err := h.userClient.RevokeSessions(r.Context(), userUUID.String(), "", []string{sessionID})
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type revokeOthersRequest struct {
	User          string `json:"user"`
	KeepSessionID string `json:"keep_session_id"`
}

func (h *AdminHandler) RevokeOthers(w http.ResponseWriter, r *http.Request) {
	var req revokeOthersRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if _, err := uuid.Parse(req.User); err != nil {
		http.Error(w, "invalid user", http.StatusBadRequest)
		return
	}
	if req.KeepSessionID == "" {
		http.Error(w, "keep_session_id is required", http.StatusBadRequest)
		return
	}
	err := h.userClient.RevokeSessions(r.Context(), req.User, req.KeepSessionID, nil)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- users ----------

func (h *AdminHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePageParams(r)
	page, err := h.userClient.ListUsers(r.Context(), limit, offset)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, page)
}

type patchUserRequest struct {
	Blocked *bool `json:"blocked"`
}

func (h *AdminHandler) PatchUser(w http.ResponseWriter, r *http.Request) {
	userUUID, ok := parseUUIDParam(w, r, "user_uuid")
	if !ok {
		return
	}
	var req patchUserRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Blocked == nil {
		http.Error(w, "blocked is required", http.StatusBadRequest)
		return
	}
	err := h.userClient.SetUserBlocked(r.Context(), userUUID.String(), *req.Blocked)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ---------- moderation (метки 18+) ----------

// explicitFilterParam: "all" (по умолчанию), "explicit", "clean".
func explicitFilterParam(r *http.Request) int32 {
	switch r.URL.Query().Get("explicit") {
	case "explicit":
		return 1
	case "clean":
		return 2
	default:
		return 0
	}
}

func (h *AdminHandler) ModerationList(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePageParams(r)
	tracks, total, err := h.catalogClient.ListTracks(r.Context(), explicitFilterParam(r), limit, offset)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"tracks": tracks, "total": total})
}

type setExplicitRequest struct {
	Explicit *bool `json:"explicit"`
}

func (h *AdminHandler) ModerationSetExplicit(w http.ResponseWriter, r *http.Request) {
	trackUUID, ok := parseUUIDParam(w, r, "track_uuid")
	if !ok {
		return
	}
	var req setExplicitRequest
	if !decodeBody(w, r, &req) {
		return
	}
	if req.Explicit == nil {
		http.Error(w, "explicit is required", http.StatusBadRequest)
		return
	}
	err := h.catalogClient.SetTrackExplicit(r.Context(), trackUUID.String(), *req.Explicit)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
