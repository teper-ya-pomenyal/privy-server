package grpc

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogWriteGRPCHandler) AddTrack(ctx context.Context, req *catalogv1.AddTrackRequest) (*catalogv1.AddTrackResponse, error) {
	if req.TrackName == "" || req.ArtistUuid == "" || req.AlbumUuid == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	artistUUID, err := uuid.Parse(req.ArtistUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	albumUUID, err := uuid.Parse(req.AlbumUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}

	trackReq := &domain.Track{
		TrackName: req.TrackName, ArtistID: artistUUID,
		AlbumID: albumUUID, Explicit: req.Explicit,
		TrackPath: req.TrackPath, CoverPath: req.CoverPath,
		DurationMS: time.Duration(req.DurationMs),
	}
	track, err := h.trackUseCase.AddTrack(ctx, trackReq)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.AddTrackResponse{
		TrackUuid:  track.TrackID.String(),
		TrackName:  track.TrackName,
		ArtistUuid: track.ArtistID.String(),
		AlbumUuid:  track.AlbumID.String(),
		Explicit:   track.Explicit,
		TrackPath:  track.TrackPath,
		CoverPath:  track.CoverPath,
		DurationMs: int32(track.DurationMS),
	}, nil
}
