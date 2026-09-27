package http

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/streaming_service/internal/client/grpc"
	"github.com/teper-ya-pomenyal/privy_stream/streaming_service/internal/usecase"
)

type Streamer interface {
	StreamTrack(tl *grpc.TrackLocation, bd time.Time) (*usecase.StreamTrackResponse, error)
}

type Catalog interface {
	GetTrackByID(ctx context.Context, trackUUID uuid.UUID) (*grpc.TrackLocation, error)
}
