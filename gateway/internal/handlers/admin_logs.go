package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/clients"
)

// AdminLogRing — кольцевой буфер HTTP-запросов gateway для GET /admin/logs.
// Пишется в дополнение к stdout-логу chi, не вместо него.
type AdminLogRing struct {
	mu      sync.Mutex
	entries []clients.LogEntry // новые в конце
	cap     int
}

func NewAdminLogRing(capacity int) *AdminLogRing {
	return &AdminLogRing{entries: make([]clients.LogEntry, 0, capacity), cap: capacity}
}

func (r *AdminLogRing) Add(e clients.LogEntry) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.entries = append(r.entries, e)
	if len(r.entries) > r.cap {
		r.entries = r.entries[len(r.entries)-r.cap:]
	}
}

// Latest возвращает последние n записей, новые первыми.
func (r *AdminLogRing) Latest(n int) []clients.LogEntry {
	r.mu.Lock()
	defer r.mu.Unlock()
	if n > len(r.entries) {
		n = len(r.entries)
	}
	out := make([]clients.LogEntry, n)
	for i, e := range r.entries[len(r.entries)-n:] {
		out[n-1-i] = e
	}
	return out
}

// AdminAccessLog оборачивает весь роутер gateway: каждый HTTP-запрос
// попадает в кольцевой буфер с уровнем по итоговому статусу. Собственный
// поллинг админки (health/logs) в буфер не пишется — иначе он вытесняет
// настоящие события из истории.
func AdminAccessLog(ring *AdminLogRing) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			started := time.Now()
			rec := &logStatusRecorder{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rec, req)

			if req.URL.Path == "/admin/health" || req.URL.Path == "/admin/logs" {
				return
			}

			ring.Add(clients.LogEntry{
				At:      time.Now().UnixMilli(),
				Level:   levelForStatus(rec.status),
				Service: "gateway",
				Method:  req.Method,
				Path:    req.URL.Path,
				Code:    strconv.Itoa(rec.status),
				MS:      time.Since(started).Milliseconds(),
			})
		})
	}
}

type logStatusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *logStatusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func levelForStatus(code int) string {
	switch {
	case code >= 500:
		return "ERROR"
	case code >= 400:
		return "WARN"
	default:
		return "INFO"
	}
}

// Logs отвечает GET /admin/logs: последние записи gateway и журналов
// user_service/catalog_service (опрос по gRPC, без SSE), merged по времени.
// streaming_service не пишет в журнал (нет gRPC-канала) — его отказы видны
// в записях gateway как 5xx на /stream.
func (h *AdminHandler) Logs(w http.ResponseWriter, r *http.Request) {
	limit := 200
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil && n > 0 {
			limit = n
		}
	}
	if limit > 500 {
		limit = 500
	}
	service := r.URL.Query().Get("service")
	level := r.URL.Query().Get("level")

	entries := h.logs.Latest(limit)
	fetchCtx := r.Context()
	if svcLogs, err := h.userClient.GetLogs(fetchCtx, int32(limit)); err == nil {
		entries = append(entries, svcLogs...)
	}
	if svcLogs, err := h.catalogClient.GetLogs(fetchCtx, int32(limit)); err == nil {
		entries = append(entries, svcLogs...)
	}

	filtered := make([]clients.LogEntry, 0, len(entries))
	for _, e := range entries {
		if service != "" && e.Service != service {
			continue
		}
		if level != "" && e.Level != level {
			continue
		}
		filtered = append(filtered, e)
	}
	sort.SliceStable(filtered, func(i, j int) bool { return filtered[i].At > filtered[j].At })
	if len(filtered) > limit {
		filtered = filtered[:limit]
	}

	writeJSON(w, http.StatusOK, map[string]any{"entries": filtered})
}
