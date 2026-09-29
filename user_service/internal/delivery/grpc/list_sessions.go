package grpc

import (
	"context"

	"github.com/google/uuid"
	userv1 "github.com/teper-ya-pomenyal/privy_stream/proto/user/v1"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (h *UserGRPCHandler) ListSessions(ctx context.Context, req *userv1.ListSessionsRequest) (*userv1.ListSessionsResponse, error) {
	if req.UserUuid == "" {
		return nil, status.Error(codes.InvalidArgument, "user_uuid is required")
	}
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	sessions, err := h.listSessionsUseCase.ListSessions(ctx, userUUID)
	if err != nil {
		return nil, mapDomainError(err)
	}
	entries := make([]*userv1.SessionEntry, 0, len(sessions))
	for _, s := range sessions {
		entries = append(entries, &userv1.SessionEntry{
			SessionId: s.SessionID,
			CreatedAt: s.CreatedAt.Unix(),
			ExpiresAt: s.ExpiresAt.Unix(),
		})
	}
	return &userv1.ListSessionsResponse{Sessions: entries}, nil
}

func (h *UserGRPCHandler) RevokeSessions(ctx context.Context, req *userv1.RevokeSessionsRequest) (*userv1.RevokeSessionsResponse, error) {
	if req.UserUuid == "" {
		return nil, status.Error(codes.InvalidArgument, "user_uuid is required")
	}
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	if req.KeepSessionId == "" && len(req.SessionIds) == 0 {
		return nil, status.Error(codes.InvalidArgument, "keep_session_id or session_ids is required")
	}
	err = h.revokeSessionsUseCase.RevokeSessions(ctx, userUUID, req.KeepSessionId, req.SessionIds)
	if err != nil {
		return nil, mapDomainError(err)
	}
	return &userv1.RevokeSessionsResponse{}, nil
}
