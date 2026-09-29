package grpc

import (
	"context"

	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
)

func (h *CatalogGRPCHandler) ListTracks(ctx context.Context, req *catalogv1.ListTracksRequest) (*catalogv1.ListTracksResponse, error) {
	tracks, total, err := h.trackUseCase.ListTracks(ctx, int(req.ExplicitFilter), int(req.Limit), int(req.Offset))
	if err != nil {
		return nil, mapDomainError(err)
	}
	respTracks := make([]*catalogv1.Track, 0, len(tracks))
	for _, t := range tracks {
		respTracks = append(respTracks, &catalogv1.Track{
			TrackUuid:  t.TrackID.String(),
			TrackName:  t.TrackName,
			ArtistUuid: t.ArtistID.String(),
			ArtistName: t.ArtistName,
			AlbumUuid:  t.AlbumID.String(),
			AlbumName:  t.AlbumName,
			Explicit:   t.Explicit,
			DurationMs: int32(t.DurationMS),
			CoverPath:  t.CoverPath,
		})
	}
	return &catalogv1.ListTracksResponse{Tracks: respTracks, Total: int32(total)}, nil
}

func (h *CatalogGRPCHandler) Health(ctx context.Context, _ *catalogv1.HealthRequest) (*catalogv1.HealthResponse, error) {
	return &catalogv1.HealthResponse{Postgres: h.healthUseCase.Check(ctx)}, nil
}
