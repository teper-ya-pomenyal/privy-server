package grpc

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/domain"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *CatalogGRPCHandler) GetAlbumByID(ctx context.Context, req *catalogv1.GetAlbumByIDRequest) (*catalogv1.GetAlbumByIDResponse, error) {
	if req.AlbumUuid == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	albumUUID, err := uuid.Parse(req.AlbumUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	album, err := h.albumUseCase.GetAlbumByID(ctx, albumUUID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &catalogv1.GetAlbumByIDResponse{
		Album: &catalogv1.Album{
			AlbumUuid:  album.AlbumUUID.String(),
			ArtistUuid: album.ArtistUUID.String(),
			AlbumName:  album.AlbumName,
			CreatedAt:  album.CreatedAt.Format(timeLayout),
			CoverPath:  album.CoverPath,
		},
	}, nil
}

func (h *CatalogGRPCHandler) GetAlbumTracks(ctx context.Context, req *catalogv1.GetAlbumTracksRequest) (*catalogv1.GetAlbumTracksResponse, error) {
	if req.AlbumUuid == "" {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidCharacters.Error())
	}
	albumUUID, err := uuid.Parse(req.AlbumUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	tracks, err := h.albumUseCase.GetAlbumTracks(ctx, albumUUID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	respTracks := make([]*catalogv1.LightTrack, 0, len(tracks))
	for _, t := range tracks {
		respTracks = append(respTracks, &catalogv1.LightTrack{
			TrackUuid:  t.TrackID.String(),
			TrackName:  t.TrackName,
			Explicit:   t.Explicit,
			DurationMs: int32(t.DurationMS),
		})
	}
	return &catalogv1.GetAlbumTracksResponse{LightTrack: respTracks}, nil
}
