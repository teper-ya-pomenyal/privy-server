package domain

import (
	"time"

	"github.com/google/uuid"
)

type Track struct {
	TrackID    uuid.UUID     `db:"track_id"`
	TrackName  string        `db:"track_name"`
	ArtistID   uuid.UUID     `db:"artist_id"`
	ArtistName string        `db:"artist_name"`
	AlbumID    uuid.UUID     `db:"album_id"`
	AlbumName  string        `db:"album_name"`
	Explicit   bool          `db:"explicit"`
	CreatedAt  time.Time     `db:"created_at"`
	TrackPath  string        `db:"path"`
	CoverPath  string        `db:"cover_path"`
	DurationMS time.Duration `db:"duration_ms"`
}

type TrackPath struct {
	Path       string        `db:"path"`
	DurationMS time.Duration `db:"duration_ms"`
	Explicit   bool          `db:"explicit"`
	AlbumID    uuid.UUID     `db:"album_id"`
	CoverPath  string        `db:"cover_path"`
}

type LightTrack struct {
	TrackID    uuid.UUID     `db:"track_id"`
	TrackName  string        `db:"track_name"`
	Explicit   bool          `db:"explicit"`
	DurationMS time.Duration `db:"duration_ms"`
}

type LightAlbumTrack struct {
	TrackID    uuid.UUID     `db:"track_id"`
	TrackName  string        `db:"track_name"`
	Explicit   bool          `db:"explicit"`
	DurationMS time.Duration `db:"duration_ms"`
	Position   int           `db:"position"`
}

type LightPlaylistTrack struct {
	TrackID    uuid.UUID     `db:"track_id"`
	TrackName  string        `db:"track_name"`
	Explicit   bool          `db:"explicit"`
	DurationMS time.Duration `db:"duration_ms"`
	Position   int           `db:"position"`
}
