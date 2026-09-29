package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"
	"time"
)

const healthTimeout = 2 * time.Second

// AdminServiceHealth — состояние одного узла инфраструктуры.
// Status: UP | DOWN. Postgres/Redis заполняются, только если проверка дошла
// до них (например, недоступный user_service не даёт ответа про redis).
type AdminServiceHealth struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	MS       int64  `json:"ms"`
	Note     string `json:"note,omitempty"`
	Postgres bool   `json:"postgres,omitempty"`
	Redis    bool   `json:"redis,omitempty"`
}

type AdminHealth struct {
	CheckedAt string               `json:"checked_at"`
	Services  []AdminServiceHealth `json:"services"`
}

// Health отвечает GET /admin/health: ping зависимостей user_service
// (postgres + redis сессий) и catalog_service (postgres), доступность
// streaming_service и хранилища треков. Проверки идут параллельно —
// ответ не ждёт самый медленный сервис.
func (h *AdminHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), healthTimeout)
	defer cancel()

	user := AdminServiceHealth{Name: "user_service"}
	catalog := AdminServiceHealth{Name: "catalog_service"}
	streaming := AdminServiceHealth{Name: "streaming_service"}
	storage := AdminServiceHealth{Name: "track_storage"}

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		started := time.Now()
		pg, redis, err := h.userClient.Health(ctx)
		user.MS = time.Since(started).Milliseconds()
		if err != nil {
			user.Status = "DOWN"
			user.Note = "gRPC недоступен"
			return
		}
		user.Postgres, user.Redis = pg, redis
		user.Status = "UP"
		user.Note = "postgres ok · redis ok"
		if !pg || !redis {
			user.Status = "DOWN"
			user.Note = fmt.Sprintf("postgres %t · redis %t", pg, redis)
		}
	}()

	go func() {
		defer wg.Done()
		started := time.Now()
		pg, err := h.catalogClient.Health(ctx)
		catalog.MS = time.Since(started).Milliseconds()
		if err != nil {
			catalog.Status = "DOWN"
			catalog.Note = "gRPC недоступен"
			return
		}
		catalog.Postgres = pg
		catalog.Status = "UP"
		catalog.Note = "postgres ok"
		if !pg {
			catalog.Status = "DOWN"
			catalog.Note = "postgres недоступен"
		}
	}()

	go func() {
		defer wg.Done()
		started := time.Now()
		note := probeStreaming(ctx, h.streamingAddr)
		streaming.MS = time.Since(started).Milliseconds()
		if note == "" {
			streaming.Status = "UP"
			streaming.Note = "проксирование /stream работает"
		} else {
			streaming.Status = "DOWN"
			streaming.Note = note
		}
	}()

	// хранилище — локальный для gateway os.Stat, быстрее сделать в том же проходе
	started := time.Now()
	if _, err := os.Stat(h.storagePath); err != nil {
		storage.Status = "DOWN"
		storage.Note = "каталог недоступен: " + h.storagePath
	} else {
		storage.Status = "UP"
		storage.Note = h.storagePath
	}
	storage.MS = time.Since(started).Milliseconds()

	wg.Wait()

	writeJSON(w, http.StatusOK, AdminHealth{
		CheckedAt: time.Now().UTC().Format(time.RFC3339),
		Services:  []AdminServiceHealth{user, catalog, streaming, storage},
	})
}

// probeStreaming спрашивает у streaming_service заведомо несуществующий трек:
// 404 значит, что сервис жив и сам сходил в каталог за ответом.
// Пустая строка — проверка пройдена.
func probeStreaming(ctx context.Context, addr string) string {
	url := fmt.Sprintf("http://%s/stream/00000000-0000-0000-0000-000000000000", addr)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "не удалось сформировать запрос"
	}
	client := &http.Client{Timeout: healthTimeout}
	res, err := client.Do(req)
	if err != nil {
		return "нет ответа"
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusNotFound || res.StatusCode == http.StatusOK || res.StatusCode == http.StatusPartialContent {
		return ""
	}
	return fmt.Sprintf("неожиданный ответ %d", res.StatusCode)
}
