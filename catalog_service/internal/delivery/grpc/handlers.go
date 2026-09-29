package grpc

import (
	"github.com/teper-ya-pomenyal/privy_stream/catalog_service/internal/usecase"
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
)

type CatalogGRPCHandler struct {
	catalogv1.UnimplementedCatalogServiceServer
	trackUseCase  *usecase.TrackUseCase
	artistUseCase *usecase.ArtistUseCase
	albumUseCase  *usecase.AlbumUseCase
	healthUseCase *usecase.HealthUseCase
}

func NewCatalogGRPCHandler(trackUseCase *usecase.TrackUseCase, artistUseCase *usecase.ArtistUseCase, albumUseCase *usecase.AlbumUseCase, healthUseCase *usecase.HealthUseCase) *CatalogGRPCHandler {
	return &CatalogGRPCHandler{
		trackUseCase:  trackUseCase,
		artistUseCase: artistUseCase,
		albumUseCase:  albumUseCase,
		healthUseCase: healthUseCase,
	}
}

type CatalogWriteGRPCHandler struct {
	catalogv1.UnimplementedCatalogWriteServiceServer
	trackUseCase  *usecase.TrackUseCase
	artistUseCase *usecase.ArtistUseCase
	albumUseCase  *usecase.AlbumUseCase
}

func NewCatalogWriteGRPCHandler(trackUseCase *usecase.TrackUseCase, artistUseCase *usecase.ArtistUseCase, albumUseCase *usecase.AlbumUseCase) *CatalogWriteGRPCHandler {
	return &CatalogWriteGRPCHandler{
		trackUseCase:  trackUseCase,
		artistUseCase: artistUseCase,
		albumUseCase:  albumUseCase,
	}
}
