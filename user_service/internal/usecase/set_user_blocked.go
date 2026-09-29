package usecase

import (
	"context"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

type SetUserBlockedUseCase struct {
	repo     domain.UsersRepository
	sessions SessionStore
}

func NewSetUserBlockedUseCase(repo domain.UsersRepository, sessions SessionStore) *SetUserBlockedUseCase {
	return &SetUserBlockedUseCase{repo: repo, sessions: sessions}
}

// SetBlocked меняет флаг блокировки. Блокировка сразу отзывает все сессии
// пользователя: иначе уже выданные refresh-токены пережили бы метку.
// Владельца узла блокировать нельзя: без него узел потерял бы управление —
// владелец не смог бы даже разблокировать себя через админку.
func (uc *SetUserBlockedUseCase) SetBlocked(ctx context.Context, userUUID uuid.UUID, blocked bool) error {
	user, err := uc.repo.GetUserByID(ctx, userUUID)
	if err != nil {
		return err
	}
	if blocked && user.Role == domain.RoleOwner {
		return domain.ErrCannotBlockOwner
	}
	if err := uc.repo.SetUserBlocked(ctx, userUUID, blocked); err != nil {
		return err
	}
	if !blocked {
		return nil
	}
	return uc.sessions.RevokeSessions(ctx, userUUID, "", nil)
}
