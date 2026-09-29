package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogWriteGRPCHandler) SetTrackExplicit(ctx context.Context, req *catalogv1.SetTrackExplicitRequest) (*catalogv1.SetTrackExplicitResponse, error) {
	if req.TrackUuid == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	trackUUID, err := uuid.Parse(req.TrackUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	err = h.trackUseCase.SetTrackExplicit(ctx, trackUUID, req.Explicit)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.SetTrackExplicitResponse{}, nil
}
