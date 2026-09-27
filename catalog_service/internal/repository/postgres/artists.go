package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
)

func (c *PostgresCatalog) SearchArtist(ctx context.Context, artistName string, limit, offset int) ([]domain.Artist, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT artist_id, artist_name
		FROM artists
		WHERE artist_name ILIKE '%' || $1 || '%'
		ORDER BY artist_name ASC
		LIMIT $2 OFFSET $3
		`,
		artistName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	artists := []domain.Artist{}
	for rows.Next() {
		var a domain.Artist
		if err := rows.Scan(
			&a.ArtistUUID, &a.ArtistName,
		); err != nil {
			return nil, err
		}
		artists = append(artists, a)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return artists, nil
}

func (c *PostgresCatalog) GetArtistAlbums(ctx context.Context, artistUUID uuid.UUID, limit, offset int32) ([]domain.Album, error) {
	rows, err := c.pool.Query(ctx, `
			SELECT album_id, album_name, created_at
			FROM albums
			WHERE artist_id = $1
			ORDER BY created_at DESC
			LIMIT $2 OFFSET $3
			`,
		artistUUID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	albums := []domain.Album{}

	for rows.Next() {
		var a domain.Album
		if err := rows.Scan(&a.AlbumUUID, &a.AlbumName, &a.CreatedAt); err != nil {
			return nil, err
		}
		albums = append(albums, a)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return albums, nil
}

func (c *PostgresCatalog) GetArtistTracks(ctx context.Context, artistUUID uuid.UUID, limit, offset int32) ([]domain.LightTrack, error) {
	rows, err := c.pool.Query(ctx, `
		SELECT track_id, track_name, explicit, duration_ms
		FROM tracks
		WHERE artist_id = $1
		ORDER BY listened DESC
		LIMIT $2 OFFSET $3
		`,
		artistUUID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	tracks := []domain.LightTrack{}

	for rows.Next() {
		t := domain.LightTrack{}
		if err := rows.Scan(&t.TrackID, &t.TrackName, &t.Explicit, &t.DurationMS); err != nil {
			return nil, err
		}
		tracks = append(tracks, t)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return tracks, nil
}

func (c *PostgresCatalog) GetArtistByID(ctx context.Context, artistUUID uuid.UUID) (*domain.Artist, error) {
	var a domain.Artist
	err := c.pool.QueryRow(ctx, `
		SELECT artist_id, artist_name
		FROM artists
		WHERE artist_id = $1
		`,
		artistUUID,
	).Scan(&a.ArtistUUID, &a.ArtistName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrArtistNotFound
		}
		return nil, err
	}
	return &a, nil
}

/////////////////////////////////////////////////////////////////

func (c *PostgresCatalog) AddArtist(ctx context.Context, artist domain.Artist) error {
	_, err := c.pool.Exec(ctx, `
		INSERT INTO artists
			(artist_id, artist_name, created_at)
		VALUES($1, $2, $3)
		`,
		artist.ArtistUUID, artist.ArtistName, artist.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrArtistAlreadyExists
		}
		return err
	}
	return nil
}
