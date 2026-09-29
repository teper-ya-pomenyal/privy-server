package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogWriteGRPCHandler) DeleteTrack(ctx context.Context, req *catalogv1.DeleteTrackRequest) (*catalogv1.DeleteTrackResponse, error) {
	if req.TrackUuid == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	trackUUID, err := uuid.Parse(req.TrackUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	trackPath, err := h.trackUseCase.DeleteTrack(ctx, trackUUID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.DeleteTrackResponse{TrackPath: trackPath.Path}, nil
}
