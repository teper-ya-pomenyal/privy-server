package grpc

import (
	"context"

	"github.com/google/uuid"
	userv1 "github.com/teper-ya-pomenyal/privy_stream/proto/user/v1"
)

func (h *UserGRPCHandler) CreateStreamToken(ctx context.Context, req *userv1.CreateStreamTokenRequest) (*userv1.CreateStreamTokenResponse, error) {
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, mapDomainError(err)
	}

	token, err := h.createStreamTokenUseCase.CreateStreamToken(userUUID, req.BirthDate.AsTime())
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &userv1.CreateStreamTokenResponse{StreamToken: token}, nil
}
