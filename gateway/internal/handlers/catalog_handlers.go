package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/clients"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/middlewares"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/ratelimit"
	"github.com/teper-ya-pomenyal/privy_stream/gateway/internal/storage"
)

const defaultPageLimit = 20
const maxTrackFileMemory = 32 << 20

type CatalogHandler struct {
	catalogClient *clients.CatalogClient
	trackStorage  *storage.TrackStorage
	listenLimiter *ratelimit.ListenLimiter
}

func NewCatalogHandler(catalogClient *clients.CatalogClient, trackStorage *storage.TrackStorage, listenLimiter *ratelimit.ListenLimiter) *CatalogHandler {
	return &CatalogHandler{catalogClient: catalogClient, trackStorage: trackStorage, listenLimiter: listenLimiter}
}

func parsePageParams(r *http.Request) (int32, int32) {
	limit := int32(defaultPageLimit)
	offset := int32(0)
	if v := r.URL.Query().Get("limit"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			limit = int32(n)
		}
	}
	if v := r.URL.Query().Get("offset"); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			offset = int32(n)
		}
	}
	return limit, offset
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func (h *CatalogHandler) GetTrackByID(w http.ResponseWriter, r *http.Request) {
	res, err := h.catalogClient.GetTrackByID(r.Context(), chi.URLParam(r, "track_uuid"))
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *CatalogHandler) TrackExists(w http.ResponseWriter, r *http.Request) {
	exists, err := h.catalogClient.TrackExists(r.Context(), chi.URLParam(r, "track_uuid"))
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"exists": exists})
}

func (h *CatalogHandler) SearchTrack(w http.ResponseWriter, r *http.Request) {
	trackName := r.URL.Query().Get("track_name")
	if trackName == "" {
		http.Error(w, "track_name is required", http.StatusBadRequest)
		return
	}
	limit, offset := parsePageParams(r)
	res, err := h.catalogClient.SearchTrack(r.Context(), trackName, limit, offset)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *CatalogHandler) GetArtistByID(w http.ResponseWriter, r *http.Request) {
	res, err := h.catalogClient.GetArtistByID(r.Context(), chi.URLParam(r, "artist_uuid"))
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *CatalogHandler) SearchArtist(w http.ResponseWriter, r *http.Request) {
	artistName := r.URL.Query().Get("artist_name")
	if artistName == "" {
		http.Error(w, "artist_name is required", http.StatusBadRequest)
		return
	}
	limit, offset := parsePageParams(r)
	res, err := h.catalogClient.SearchArtist(r.Context(), artistName, limit, offset)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *CatalogHandler) GetArtistAlbums(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePageParams(r)
	res, err := h.catalogClient.GetArtistAlbums(r.Context(), chi.URLParam(r, "artist_uuid"), limit, offset)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *CatalogHandler) GetArtistTracks(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePageParams(r)
	res, err := h.catalogClient.GetArtistTracks(r.Context(), chi.URLParam(r, "artist_uuid"), limit, offset)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

func (h *CatalogHandler) GetAlbumByID(w http.ResponseWriter, r *http.Request) {
	res, err := h.catalogClient.GetAlbumByID(r.Context(), chi.URLParam(r, "album_uuid"))
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

// Content-Type обложки по расширению. Файлы пишутся только с расширениями из
// allowedImageExt, так что таблица полная; неизвестное расширение не переопределяет
// автоопределение http.ServeContent.
var coverContentTypes = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".webp": "image/webp",
	".gif":  "image/gif",
}

// GetAlbumCover отдаёт файл обложки альбома (cover_path из метаданных) — вторая
// половина фичи после POST .../cover. <img> не отправляет Authorization, поэтому
// клиент забирает картинку авторизованным fetch, как и /stream.
func (h *CatalogHandler) GetAlbumCover(w http.ResponseWriter, r *http.Request) {
	albumUUID, err := uuid.Parse(chi.URLParam(r, "album_uuid"))
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	album, err := h.catalogClient.GetAlbumByID(r.Context(), albumUUID.String())
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	if album.CoverPath == "" {
		http.Error(w, "album has no cover", http.StatusNotFound)
		return
	}
	file, info, err := h.trackStorage.Open(album.CoverPath)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidPath) {
			log.Printf("album %s has unsafe cover path %q, rejected", albumUUID, album.CoverPath)
			http.Error(w, "invalid cover path", http.StatusBadRequest)
			return
		}
		http.Error(w, "cover file not found", http.StatusNotFound)
		return
	}
	defer file.Close()

	if ct, ok := coverContentTypes[strings.ToLower(filepath.Ext(album.CoverPath))]; ok {
		w.Header().Set("Content-Type", ct)
	}
	// Имя файла генерируется заново при каждой загрузке, так что содержимое по
	// одному пути не меняется — можно кэшировать.
	w.Header().Set("Cache-Control", "private, max-age=3600")
	http.ServeContent(w, r, filepath.Base(album.CoverPath), info.ModTime(), file)
}

func (h *CatalogHandler) GetAlbumTracks(w http.ResponseWriter, r *http.Request) {
	res, err := h.catalogClient.GetAlbumTracks(r.Context(), chi.URLParam(r, "album_uuid"))
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, res)
}

type AddArtistRequest struct {
	ArtistName string `json:"artist_name"`
}

func (h *CatalogHandler) AddArtist(w http.ResponseWriter, r *http.Request) {
	var req AddArtistRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.ArtistName == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	res, err := h.catalogClient.AddArtist(r.Context(), req.ArtistName)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

type AddAlbumRequest struct {
	ArtistUUID string `json:"artist_uuid"`
	AlbumName  string `json:"album_name"`
}

func (h *CatalogHandler) AddAlbum(w http.ResponseWriter, r *http.Request) {
	var req AddAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.ArtistUUID == "" || req.AlbumName == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	res, err := h.catalogClient.AddAlbum(r.Context(), req.ArtistUUID, req.AlbumName)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

type AddTrackRequest struct {
	TrackName  string `json:"track_name"`
	ArtistUUID string `json:"artist_uuid"`
	AlbumUUID  string `json:"album_uuid"`
	Explicit   bool   `json:"explicit"`
	Path       string `json:"path"`
	DurationMs int32  `json:"duration_ms"`
}

func (h *CatalogHandler) AddTrack(w http.ResponseWriter, r *http.Request) {
	var req AddTrackRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.TrackName == "" || req.ArtistUUID == "" || req.AlbumUUID == "" {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	albumUUID, err := uuid.Parse(req.AlbumUUID)
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	// Путь к файлу задаёт сервер: клиентский req.Path используется только как подсказка расширения.
	trackPath := storage.NewTrackPath(albumUUID, req.Path)
	res, err := h.catalogClient.AddTrack(r.Context(), req.TrackName, req.ArtistUUID, albumUUID.String(), req.Explicit, trackPath, req.DurationMs)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, res)
}

type AddTracksToAlbumRequest struct {
	Tracks []clients.AlbumTrackInput `json:"tracks"`
}

func (h *CatalogHandler) AddTracksToAlbum(w http.ResponseWriter, r *http.Request) {
	var req AddTracksToAlbumRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if len(req.Tracks) == 0 {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	albumUUID := chi.URLParam(r, "album_uuid")
	if err := h.catalogClient.AddTracksToAlbum(r.Context(), albumUUID, req.Tracks); err != nil {
		mapGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (h *CatalogHandler) AddTrackFile(w http.ResponseWriter, r *http.Request) {
	trackUUID := chi.URLParam(r, "track_uuid")
	track, err := h.catalogClient.GetTrackByID(r.Context(), trackUUID)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	if track.Path == "" {
		http.Error(w, "track has no path set", http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(maxTrackFileMemory); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	defer file.Close()

	size, err := h.trackStorage.AddFile(track.Path, file)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidPath) {
			log.Printf("track %s has unsafe path %q, upload rejected", trackUUID, track.Path)
			http.Error(w, "invalid track path", http.StatusBadRequest)
			return
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"path": track.Path, "size": size})
}

// uploadCover принимает multipart-файл, генерирует серверный путь обложки
// по тому же принципу, что и путь трека, и записывает файл в хранилище.
// Файл пишется до обновления каталога: осиротевший файл безопаснее,
// чем ссылка на несуществующий файл в БД.
func (h *CatalogHandler) uploadCover(w http.ResponseWriter, r *http.Request, albumUUID uuid.UUID) (string, int64, bool) {
	if err := r.ParseMultipartForm(maxTrackFileMemory); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return "", 0, false
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return "", 0, false
	}
	defer file.Close()

	coverPath := storage.NewCoverPath(albumUUID, header.Filename)
	size, err := h.trackStorage.AddFile(coverPath, file)
	if err != nil {
		if errors.Is(err, storage.ErrInvalidPath) {
			http.Error(w, "invalid cover path", http.StatusBadRequest)
			return "", 0, false
		}
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return "", 0, false
	}
	return coverPath, size, true
}

func (h *CatalogHandler) AddTrackCover(w http.ResponseWriter, r *http.Request) {
	trackUUID := chi.URLParam(r, "track_uuid")
	track, err := h.catalogClient.GetTrackByID(r.Context(), trackUUID)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	albumUUID, err := uuid.Parse(track.AlbumUUID)
	if err != nil {
		http.Error(w, "track has no valid album", http.StatusBadRequest)
		return
	}

	coverPath, size, ok := h.uploadCover(w, r, albumUUID)
	if !ok {
		return
	}
	if err := h.catalogClient.SetTrackCover(r.Context(), trackUUID, coverPath); err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"path": coverPath, "size": size})
}

func (h *CatalogHandler) AddAlbumCover(w http.ResponseWriter, r *http.Request) {
	albumUUID, err := uuid.Parse(chi.URLParam(r, "album_uuid"))
	if err != nil {
		http.Error(w, "invalid uuid", http.StatusBadRequest)
		return
	}
	if _, err := h.catalogClient.GetAlbumByID(r.Context(), albumUUID.String()); err != nil {
		mapGRPCError(w, err)
		return
	}

	coverPath, size, ok := h.uploadCover(w, r, albumUUID)
	if !ok {
		return
	}
	if err := h.catalogClient.SetAlbumCover(r.Context(), albumUUID.String(), coverPath); err != nil {
		mapGRPCError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"path": coverPath, "size": size})
}

// DeleteTrack удаляет трек из каталога и его файл из хранилища. Файл убирается
// после записи в БД: осиротевший файл безопаснее ссылки на несуществующий файл.
func (h *CatalogHandler) DeleteTrack(w http.ResponseWriter, r *http.Request) {
	trackUUID := chi.URLParam(r, "track_uuid")
	path, err := h.catalogClient.DeleteTrack(r.Context(), trackUUID)
	if err != nil {
		mapGRPCError(w, err)
		return
	}
	if err := h.trackStorage.Remove(path); err != nil {
		log.Printf("track %s: remove file %q: %v", trackUUID, path, err)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *CatalogHandler) IncrementListened(w http.ResponseWriter, r *http.Request) {
	trackUUID := chi.URLParam(r, "track_uuid")
	userID, ok := middlewares.UserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "missing or invalid authorization header", http.StatusUnauthorized)
		return
	}

	// Если Redis недоступен, прослушивание всё равно засчитываем (fail-open):
	// счётчик не критичен, а отказ ломал бы клиенту статистику целиком.
	allowed, err := h.listenLimiter.Allow(r.Context(), userID, trackUUID)
	if err != nil {
		log.Printf("listen rate limiter unavailable: %v", err)
		allowed = true
	}
	if !allowed {
		http.Error(w, "listen already counted recently", http.StatusTooManyRequests)
		return
	}

	if err := h.catalogClient.IncrementListened(r.Context(), trackUUID); err != nil {
		if relErr := h.listenLimiter.Release(context.WithoutCancel(r.Context()), userID, trackUUID); relErr != nil {
			log.Printf("listen rate limiter release: %v", relErr)
		}
		mapGRPCError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
