package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/utils"
)

type TrackUseCase struct {
	repo CatalogRepository
}

func NewTrackUserCase(repo CatalogRepository) *TrackUseCase {
	return &TrackUseCase{repo: repo}
}

func (t *TrackUseCase) GetTrackByID(ctx context.Context, trackUUID uuid.UUID) (*domain.TrackPath, error) {
	trackPath, err := t.repo.GetTrackByID(ctx, trackUUID)
	if err != nil {
		return nil, err
	}
	return trackPath, nil
}

func (t *TrackUseCase) TrackExists(ctx context.Context, trackUUID uuid.UUID) (bool, error) {
	ok, err := t.repo.TrackExists(ctx, trackUUID)
	return ok, err
}

func (t *TrackUseCase) SearchTracks(ctx context.Context, trackName string, limit, offset int) ([]domain.Track, error) {
	if limit < 1 || offset < 0 {
		return nil, domain.ErrInvalidPageParameters
	}
	cleanTN, err := utils.ValidateTrackName(trackName)
	if err != nil {
		return nil, err
	}
	tracks, err := t.repo.SearchTrack(ctx, cleanTN, limit, offset)
	if err != nil {
		return nil, err
	}
	return tracks, nil
}

////////////////////////////////////////////////////

func (t *TrackUseCase) AddTrack(ctx context.Context, track *domain.Track) (*domain.Track, error) {
	cleanTN, err := utils.ValidateTrackName(track.TrackName)
	if err != nil {
		return nil, err
	}

	track.TrackName = cleanTN
	track.TrackID = uuid.New()
	track.CreatedAt = time.Now()
	if err := t.repo.AddTrack(ctx, track); err != nil {
		return nil, err
	}
	return track, nil
}

func (t *TrackUseCase) IncrementListened(ctx context.Context, trackUUID uuid.UUID) error {
	return t.repo.IncrementListened(ctx, trackUUID)
}

func (t *TrackUseCase) SetTrackCover(ctx context.Context, trackUUID uuid.UUID, coverPath string) error {
	return t.repo.SetTrackCover(ctx, trackUUID, coverPath)
}

func (t *TrackUseCase) DeleteTrack(ctx context.Context, trackUUID uuid.UUID) (*domain.TrackPath, error) {
	return t.repo.DeleteTrack(ctx, trackUUID)
}

// Admin: страница треков каталога для модерации меток.
func (t *TrackUseCase) ListTracks(ctx context.Context, explicitFilter, limit, offset int) ([]domain.Track, int, error) {
	if explicitFilter < 0 || explicitFilter > 2 {
		return nil, 0, domain.ErrInvalidPageParameters
	}
	if limit < 1 || offset < 0 {
		return nil, 0, domain.ErrInvalidPageParameters
	}
	return t.repo.ListTracks(ctx, explicitFilter, limit, offset)
}

func (t *TrackUseCase) SetTrackExplicit(ctx context.Context, trackUUID uuid.UUID, explicit bool) error {
	return t.repo.SetTrackExplicit(ctx, trackUUID, explicit)
}
