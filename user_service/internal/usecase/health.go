package usecase

import (
	"context"

	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

type HealthStatus struct {
	Postgres bool
	Redis    bool
}

type HealthUseCase struct {
	repo     domain.UsersRepository
	sessions SessionStore
}

func NewHealthUseCase(repo domain.UsersRepository, sessions SessionStore) *HealthUseCase {
	return &HealthUseCase{repo: repo, sessions: sessions}
}

// Check возвращает состояние зависимостей как данные, а не ошибку:
// недоступный postgres/redis — это деградация сервиса, о которой отвечает health.
func (uc *HealthUseCase) Check(ctx context.Context) HealthStatus {
	return HealthStatus{
		Postgres: uc.repo.Ping(ctx) == nil,
		Redis:    uc.sessions.Ping(ctx) == nil,
	}
}
