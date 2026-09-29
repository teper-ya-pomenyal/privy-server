package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogWriteGRPCHandler) AddAlbum(ctx context.Context, req *catalogv1.AddAlbumRequest) (*catalogv1.AddAlbumResponse, error) {
	if req.ArtistUuid == "" || req.AlbumName == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	artistUUID, err := uuid.Parse(req.ArtistUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	albumReq := &domain.Album{
		ArtistUUID: artistUUID,
		AlbumName:  req.AlbumName,
		CoverPath:  req.CoverPath,
	}
	album, err := h.albumUseCase.AddAlbum(ctx, albumReq)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.AddAlbumResponse{
		Album: &catalogv1.Album{
			AlbumUuid:  album.AlbumUUID.String(),
			ArtistUuid: album.ArtistUUID.String(),
			AlbumName:  album.AlbumName,
			CreatedAt:  album.CreatedAt.Format(timeLayout),
			CoverPath:  album.CoverPath,
		},
	}, nil
}
