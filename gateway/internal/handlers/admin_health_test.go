package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// Проба стриминга обязана идти с X-Birth-Date: streaming_service разбирает
// заголовок до обращения в каталог и отвечает 400 на пустой — без него
// health вечно показывал бы DOWN на живом сервисе.
func TestProbeStreamingSendsBirthDate(t *testing.T) {
	var sawHeader string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		sawHeader = r.Header.Get("X-Birth-Date")
		if _, err := time.Parse(time.DateOnly, sawHeader); err != nil {
			http.Error(w, "invalid birth date", http.StatusBadRequest)
			return
		}
		http.Error(w, "track not found", http.StatusNotFound)
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	if note := probeStreaming(context.Background(), addr); note != "" {
		t.Fatalf("probe of a live streaming service must pass, got note %q", note)
	}
}

func TestProbeStreamingReportsDown(t *testing.T) {
	// сервер, который никогда не отвечает живым кодом
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()

	addr := strings.TrimPrefix(srv.URL, "http://")
	note := probeStreaming(context.Background(), addr)
	if note == "" {
		t.Fatal("probe must report a failure note for 500 answers")
	}
}
