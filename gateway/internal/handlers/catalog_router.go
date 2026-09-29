package handlers

import (
	"github.com/go-chi/chi/v5"
	mw "github.com/teper-ya-pomenyal/privy_stream/gateway/internal/middlewares"
)

func (h *CatalogHandler) MountRoutes(r chi.Router, m *mw.MiddleWares) {
	r.Route("/catalog", func(r chi.Router) {
		r.Use(m.Auth.Handle)

		r.Get("/tracks/search", h.SearchTrack)
		r.Get("/tracks/{track_uuid}", h.GetTrackByID)
		r.Get("/tracks/{track_uuid}/exists", h.TrackExists)
		r.Post("/tracks/{track_uuid}/listened", h.IncrementListened)
		r.Delete("/tracks/{track_uuid}", h.DeleteTrack)

		r.Get("/artists/search", h.SearchArtist)
		r.Get("/artists/{artist_uuid}", h.GetArtistByID)
		r.Get("/artists/{artist_uuid}/albums", h.GetArtistAlbums)
		r.Get("/artists/{artist_uuid}/tracks", h.GetArtistTracks)
		r.Post("/artists", h.AddArtist)

		r.Get("/albums/{album_uuid}", h.GetAlbumByID)
		r.Get("/albums/{album_uuid}/cover", h.GetAlbumCover)
		r.Get("/albums/{album_uuid}/tracks", h.GetAlbumTracks)
		r.Post("/albums", h.AddAlbum)
		r.Post("/albums/{album_uuid}/tracks", h.AddTracksToAlbum)
		r.Post("/albums/{album_uuid}/cover", h.AddAlbumCover)
		r.Delete("/albums/{album_uuid}", h.DeleteAlbum)

		r.Post("/tracks", h.AddTrack)
		r.Post("/tracks/{track_uuid}/file", h.AddTrackFile)
		r.Post("/tracks/{track_uuid}/cover", h.AddTrackCover)
	})
}
