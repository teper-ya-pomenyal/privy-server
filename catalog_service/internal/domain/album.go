package domain

import (
	"time"

	"github.com/google/uuid"
)

type Album struct {
	AlbumUUID  uuid.UUID `db:"album_id"`
	ArtistUUID uuid.UUID `db:"artist_id"`
	AlbumName  string    `db:"album_name"`
	CreatedAt  time.Time `db:"created_at"`
	CoverPath  string    `db:"cover_path"`
}
