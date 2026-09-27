package grpc

import (
	catalogv1 "github.com/teper-ya-pomenyal/privy_stream/proto/catalog/v1"
	"google.golang.org/grpc"
)

type CatalogClient struct {
	client catalogv1.CatalogServiceClient
}

func NewCatalogClient(conn *grpc.ClientConn) *CatalogClient {
	return &CatalogClient{
		client: catalogv1.NewCatalogServiceClient(conn),
	}
}

type TrackLocation struct {
	Path       string
	DurationMS int32
	Explicit   bool
}
