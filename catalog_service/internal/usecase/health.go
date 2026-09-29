package usecase

import (
	"context"
)

type HealthUseCase struct {
	repo CatalogRepository
}

func NewHealthUseCase(repo CatalogRepository) *HealthUseCase {
	return &HealthUseCase{repo: repo}
}

// Check сообщает доступность postgres как данные: недоступная БД —
// деградация сервиса, о которой отвечает health, а не ошибка вызова.
func (uc *HealthUseCase) Check(ctx context.Context) bool {
	return uc.repo.Ping(ctx) == nil
}
