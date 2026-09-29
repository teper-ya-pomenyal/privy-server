package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogWriteGRPCHandler) DeleteAlbum(ctx context.Context, req *catalogv1.DeleteAlbumRequest) (*catalogv1.DeleteAlbumResponse, error) {
	if req.AlbumUuid == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	albumUUID, err := uuid.Parse(req.AlbumUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	files, err := h.albumUseCase.DeleteAlbum(ctx, albumUUID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.DeleteAlbumResponse{TrackPaths: files.TrackPaths, CoverPath: files.CoverPath}, nil
}
