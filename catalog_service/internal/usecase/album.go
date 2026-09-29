package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/utils"
)

type AlbumUseCase struct {
	repo CatalogRepository
}

func NewAlbumUseCase(repo CatalogRepository) *AlbumUseCase {
	return &AlbumUseCase{repo: repo}
}

func (a *AlbumUseCase) GetAlbumByID(ctx context.Context, albumUUID uuid.UUID) (*domain.Album, error) {
	album, err := a.repo.GetAlbumByID(ctx, albumUUID)
	if err != nil {
		return nil, err
	}
	return album, nil
}

func (a *AlbumUseCase) GetAlbumTracks(ctx context.Context, albumUUID uuid.UUID) ([]domain.LightAlbumTrack, error) {
	tracks, err := a.repo.GetAlbumTracks(ctx, albumUUID)
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

///////////////////////////////////////

func (a *AlbumUseCase) AddAlbum(ctx context.Context, album *domain.Album) (*domain.Album, error) {
	cleanAlN, err := utils.ValidateAlbumName(album.AlbumName)
	if err != nil {
		return nil, err
	}

	album.AlbumName = cleanAlN
	album.AlbumUUID = uuid.New()
	album.CreatedAt = time.Now()
	if err := a.repo.AddAlbum(ctx, album); err != nil {
		return nil, err
	}
	return album, nil
}

func (a *AlbumUseCase) AddTracksToAlbum(ctx context.Context, tracks []domain.AlbumTrack) error {
	return a.repo.AddTracksToAlbum(ctx, tracks)
}

func (a *AlbumUseCase) SetAlbumCover(ctx context.Context, albumUUID uuid.UUID, coverPath string) error {
	return a.repo.SetAlbumCover(ctx, albumUUID, coverPath)
}

func (a *AlbumUseCase) DeleteAlbum(ctx context.Context, albumUUID uuid.UUID) (*domain.AlbumFiles, error) {
	return a.repo.DeleteAlbum(ctx, albumUUID)
}
