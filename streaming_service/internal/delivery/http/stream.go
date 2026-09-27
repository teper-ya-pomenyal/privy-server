package http

import (
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/streaming_service/internal/domain"
)

func (h *HTTPHandlers) StreamTrack(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	trackUUID, err := uuid.Parse(id)
	if err != nil {
		http.Error(w, "invalid track-id", http.StatusBadRequest)
		log.Println(err)
		return
	}
	bdHeader := r.Header.Get("X-Birth-Date")
	bd, err := time.Parse(time.DateOnly, bdHeader)
	if err != nil {
		http.Error(w, "invalid birth date", http.StatusBadRequest)
		log.Println(err)
		return
	}

	tl, err := h.catalog.GetTrackByID(r.Context(), trackUUID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrTrackNotFound):
			http.Error(w, "track not found", http.StatusNotFound)
			log.Println(err)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}
	track, err := h.streamer.StreamTrack(tl, bd)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrExplicitContentBlocked):
			http.Error(w, "explicit content blocked", http.StatusForbidden)
			log.Println(err)
			return
		case errors.Is(err, domain.ErrFileNotExists):
			http.Error(w, "track file not found", http.StatusNotFound)
			log.Println(err)
			return
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
			log.Println(err)
			return
		}
	}
	defer track.File.Close()

	http.ServeContent(w, r, track.FileName, track.ModTime, track.File)

}
