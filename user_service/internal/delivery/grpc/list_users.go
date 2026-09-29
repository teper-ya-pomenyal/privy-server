package grpc

import (
	"context"
	"errors"

	"github.com/google/uuid"
	userv1 "github.com/teper-ya-pomenyal/privy_stream/proto/user/v1"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const (
	maxUsersPageLimit = 200
	defaultUsersLimit = 50
)

func (h *UserGRPCHandler) ListUsers(ctx context.Context, req *userv1.ListUsersRequest) (*userv1.ListUsersResponse, error) {
	// пагинация — формат входных данных, проверяется в хендлере
	limit := int(req.Limit)
	if limit == 0 {
		limit = defaultUsersLimit
	}
	if limit < 1 || limit > maxUsersPageLimit || req.Offset < 0 {
		return nil, status.Error(codes.InvalidArgument, "invalid pagination parameters")
	}

	res, err := h.listUsersUseCase.ListUsers(ctx, limit, int(req.Offset))
	if err != nil {
		return nil, mapDomainError(err)
	}
	users := make([]*userv1.UserEntry, 0, len(res.Users))
	for _, u := range res.Users {
		users = append(users, &userv1.UserEntry{
			UserUuid:  u.UserUUID.String(),
			UserName:  u.UserName,
			Role:      u.Role,
			Blocked:   u.Blocked,
			BirthDate: timestamppb.New(u.BirthDate),
			CreatedAt: timestamppb.New(u.CreatedAt),
		})
	}
	return &userv1.ListUsersResponse{Users: users, Total: int32(res.Total)}, nil
}

func (h *UserGRPCHandler) SetUserBlocked(ctx context.Context, req *userv1.SetUserBlockedRequest) (*userv1.SetUserBlockedResponse, error) {
	if req.UserUuid == "" {
		return nil, status.Error(codes.InvalidArgument, "user_uuid is required")
	}
	userUUID, err := uuid.Parse(req.UserUuid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, domain.ErrInvalidUUID.Error())
	}
	err = h.setUserBlockedUseCase.SetBlocked(ctx, userUUID, req.Blocked)
	if err != nil {
		// В админ-операции нет логина: несуществующий аккаунт — 404, а не
		// Unauthenticated (иначе gateway вернёт 401 и админка разлогинит владельца).
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, status.Error(codes.NotFound, err.Error())
		}
		return nil, mapDomainError(err)
	}
	return &userv1.SetUserBlockedResponse{}, nil
}

func (h *UserGRPCHandler) Health(ctx context.Context, _ *userv1.HealthRequest) (*userv1.HealthResponse, error) {
	health := h.healthUseCase.Check(ctx)
	return &userv1.HealthResponse{Postgres: health.Postgres, Redis: health.Redis}, nil
}
