package usecase

import (
	"context"

	"github.com/google/uuid"
)

type RevokeSessionsUseCase struct {
	sessions SessionStore
}

func NewRevokeSessionsUseCase(sessions SessionStore) *RevokeSessionsUseCase {
	return &RevokeSessionsUseCase{sessions: sessions}
}

// RevokeSessions отзывает перечисленные сессии; при пустом списке — все,
// кроме keepSessionID (пустой keep — все сессии пользователя).
func (uc *RevokeSessionsUseCase) RevokeSessions(ctx context.Context, userUUID uuid.UUID, keepSessionID string, remove []string) error {
	return uc.sessions.RevokeSessions(ctx, userUUID, keepSessionID, remove)
}
