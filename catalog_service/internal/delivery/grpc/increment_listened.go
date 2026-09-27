package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogWriteGRPCHandler) IncrementListened(ctx context.Context, req *catalogv1.IncrementListenedRequest) (*catalogv1.IncrementListenedResponse, error) {
	trackUUID, err := uuid.Parse(req.TrackUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	if err := h.trackUseCase.IncrementListened(ctx, trackUUID); err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.IncrementListenedResponse{}, nil
}
