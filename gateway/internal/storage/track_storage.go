package storage

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

// ErrInvalidPath — путь трека выходит за пределы хранилища или пустой.
var ErrInvalidPath = errors.New("invalid track path")

var allowedAudioExt = map[string]bool{
	".mp3":  true,
	".flac": true,
	".ogg":  true,
	".opus": true,
	".m4a":  true,
	".aac":  true,
	".wav":  true,
}

var allowedImageExt = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".webp": true,
	".gif":  true,
}

type TrackStorage struct {
	basePath string
}

func NewTrackStorage(basePath string) *TrackStorage {
	return &TrackStorage{basePath: basePath}
}

// NewTrackPath генерирует путь файла трека на стороне сервера: {album_uuid}/{random_uuid}{ext}.
// Из пути, присланного клиентом, берётся только расширение, и только из белого списка:
// по нему streaming определяет Content-Type. Неизвестное расширение отбрасывается,
// тогда Content-Type определяется по содержимому файла.
func NewTrackPath(albumUUID uuid.UUID, clientPath string) string {
	ext := strings.ToLower(filepath.Ext(clientPath))
	if !allowedAudioExt[ext] {
		ext = ""
	}
	return albumUUID.String() + "/" + uuid.NewString() + ext
}

// NewCoverPath генерирует путь файла обложки по тому же принципу, что и трек:
// {album_uuid}/{random_uuid}{ext}. Из присланного клиентом имени берётся только
// расширение, и только из белого списка. Файл один, ссылку на него получают
// и альбом, и его треки.
func NewCoverPath(albumUUID uuid.UUID, clientPath string) string {
	ext := strings.ToLower(filepath.Ext(clientPath))
	if !allowedImageExt[ext] {
		ext = ""
	}
	return albumUUID.String() + "/" + uuid.NewString() + ext
}

// AddFile записывает файл по пути, сгенерированному на сервере
// (NewTrackPath / NewCoverPath).
func (s *TrackStorage) AddFile(path string, r io.Reader) (int64, error) {
	// Защита для треков, созданных до серверной генерации путей: в базе может лежать
	// путь вида "../../что-угодно".
	if !filepath.IsLocal(path) {
		return 0, ErrInvalidPath
	}
	fullPath := filepath.Join(s.basePath, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return 0, err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	return io.Copy(file, r)
}
