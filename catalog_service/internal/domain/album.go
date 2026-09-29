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

// AlbumFiles — файлы хранилища, ссылки на которые ушли вместе с альбомом:
// файлы его треков и файл обложки. Удаляет их gateway, у catalog_service
// тома хранилища нет.
type AlbumFiles struct {
	TrackPaths []string
	CoverPath  string
}
