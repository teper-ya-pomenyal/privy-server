package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogWriteGRPCHandler) SetAlbumCover(ctx context.Context, req *catalogv1.SetAlbumCoverRequest) (*catalogv1.SetAlbumCoverResponse, error) {
	if req.AlbumUuid == "" || req.CoverPath == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	albumUUID, err := uuid.Parse(req.AlbumUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	if err := h.albumUseCase.SetAlbumCover(ctx, albumUUID, req.CoverPath); err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.SetAlbumCoverResponse{}, nil
}
