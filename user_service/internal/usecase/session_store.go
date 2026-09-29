package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

type SessionStore interface {
	Save(ctx context.Context, refreshToken string, userUUID uuid.UUID) error
	Refresh(ctx context.Context, oldToken, newToken string) error
	Get(ctx context.Context, refreshToken string) (uuid.UUID, error)
	Delete(ctx context.Context, refreshToken string) error
	// Admin: список активных сессий и отзыв по id (SHA-256 токена).
	ListSessions(ctx context.Context, userUUID uuid.UUID) ([]domain.SessionInfo, error)
	RevokeSessions(ctx context.Context, userUUID uuid.UUID, keepSessionID string, remove []string) error
	Ping(ctx context.Context) error
}
