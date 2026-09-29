package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
)

// SetAlbumCover записывает путь обложки альбому и всем его трекам:
// обложка трека всегда совпадает с обложкой альбома (у альбома приоритет).
func (c *PostgresCatalog) SetAlbumCover(ctx context.Context, albumUUID uuid.UUID, coverPath string) error {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	res, err := tx.Exec(ctx, `
		UPDATE albums
		SET cover_path = $2
		WHERE album_id = $1
		`, albumUUID, coverPath,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrAlbumNotFound
	}

	if _, err := tx.Exec(ctx, `
		UPDATE tracks
		SET cover_path = $2
		WHERE album_id = $1
		`, albumUUID, coverPath,
	); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// SetTrackCover ставит обложку трека через его альбом: файл один,
// ссылку получает и альбом, и все его треки.
func (c *PostgresCatalog) SetTrackCover(ctx context.Context, trackUUID uuid.UUID, coverPath string) error {
	var albumID uuid.UUID
	err := c.pool.QueryRow(ctx, `
		SELECT album_id
		FROM tracks
		WHERE track_id = $1
		`, trackUUID,
	).Scan(&albumID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrTrackNotFound
		}
		return err
	}

	return c.SetAlbumCover(ctx, albumID, coverPath)
}
