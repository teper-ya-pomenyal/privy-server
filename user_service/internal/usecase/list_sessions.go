package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

type ListSessionsUseCase struct {
	sessions SessionStore
}

func NewListSessionsUseCase(sessions SessionStore) *ListSessionsUseCase {
	return &ListSessionsUseCase{sessions: sessions}
}

func (uc *ListSessionsUseCase) ListSessions(ctx context.Context, userUUID uuid.UUID) ([]domain.SessionInfo, error) {
	return uc.sessions.ListSessions(ctx, userUUID)
}
