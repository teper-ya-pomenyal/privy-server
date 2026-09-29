package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogWriteGRPCHandler) SetTrackCover(ctx context.Context, req *catalogv1.SetTrackCoverRequest) (*catalogv1.SetTrackCoverResponse, error) {
	if req.TrackUuid == "" || req.CoverPath == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	trackUUID, err := uuid.Parse(req.TrackUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	if err := h.trackUseCase.SetTrackCover(ctx, trackUUID, req.CoverPath); err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.SetTrackCoverResponse{}, nil
}
