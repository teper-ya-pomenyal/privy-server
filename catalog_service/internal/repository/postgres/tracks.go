package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
)

func (c *PostgresCatalog) GetTrackByID(ctx context.Context, trackUUID uuid.UUID) (*domain.TrackPath, error) {
	trackPath := &domain.TrackPath{}

	err := c.pool.QueryRow(ctx, `
		SELECT path, duration_ms, explicit, album_id, COALESCE(cover_path, '')
		FROM tracks
		WHERE track_id = $1
		`,
		trackUUID,
	).Scan(&trackPath.Path, &trackPath.DurationMS, &trackPath.Explicit, &trackPath.AlbumID, &trackPath.CoverPath)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTrackNotFound
		}
		return nil, err
	}
	return trackPath, nil
}

func (c *PostgresCatalog) SearchTrack(ctx context.Context, trackName string, limit, offset int) ([]domain.Track, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT
			t.track_id, t.track_name, t.artist_id, ar.artist_name,
			COALESCE(t.cover_path, ''), t.album_id,
			al.album_name, t.explicit, t.created_at, t.duration_ms
		FROM tracks t
		JOIN artists ar ON ar.artist_id = t.artist_id
		JOIN albums al ON al.album_id = t.album_id
		WHERE t.track_name ILIKE '%' || $1 || '%' ESCAPE '\'
		ORDER BY t.listened DESC
		LIMIT $2 OFFSET $3
		`,
		trackName, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tracks := []domain.Track{}
	for rows.Next() {
		var t domain.Track
		if err := rows.Scan(
			&t.TrackID, &t.TrackName, &t.ArtistID, &t.ArtistName, &t.CoverPath,
			&t.AlbumID, &t.AlbumName, &t.Explicit, &t.CreatedAt, &t.DurationMS,
		); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tracks, nil

}

func (c *PostgresCatalog) TrackExists(ctx context.Context, trackUUID uuid.UUID) (bool, error) {
	var exists bool
	err := c.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM tracks WHERE track_id = $1)
		`,
		trackUUID,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

//////////////////////////////////////////////////////

func (c *PostgresCatalog) AddTrack(ctx context.Context, track *domain.Track) error {
	_, err := c.pool.Exec(ctx,
		`INSERT INTO tracks
		(track_id, track_name, artist_id, album_id,
		explicit, created_at, path, duration_ms, cover_path)
	 VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		track.TrackID, track.TrackName, track.ArtistID,
		track.AlbumID, track.Explicit, track.CreatedAt,
		track.TrackPath, track.DurationMS, track.CoverPath,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrTrackAlreadyExists
		}
		return err
	}

	return nil
}

func (c *PostgresCatalog) IncrementListened(ctx context.Context, trackUUID uuid.UUID) error {
	res, err := c.pool.Exec(ctx, `
		UPDATE tracks
		SET listened = listened + 1
		WHERE track_id = $1
		`,
		trackUUID,
	)
	if err != nil {
		return err
	}

	n := res.RowsAffected()

	if n == 0 {
		return domain.ErrTrackNotFound
	}
	return nil
}

// DeleteTrack удаляет трек вместе с вхождениями в трек-листы и плейлисты —
// внешние ключи на tracks созданы без CASCADE. Возвращает путь файла,
// чтобы вызывающий убрал его из хранилища; обложка не трогается —
// файл обложки принадлежит альбому и его остальным трекам.
func (c *PostgresCatalog) DeleteTrack(ctx context.Context, trackUUID uuid.UUID) (*domain.TrackPath, error) {
	tx, err := c.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	if _, err = tx.Exec(ctx, `DELETE FROM playlists_tracks WHERE track_id = $1`, trackUUID); err != nil {
		return nil, err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM albums_tracks WHERE track_id = $1`, trackUUID); err != nil {
		return nil, err
	}

	var path string
	err = tx.QueryRow(ctx, `
		DELETE FROM tracks
		WHERE track_id = $1
		RETURNING path
		`,
		trackUUID,
	).Scan(&path)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTrackNotFound
		}
		return nil, err
	}

	return &domain.TrackPath{Path: path}, tx.Commit(ctx)
}
