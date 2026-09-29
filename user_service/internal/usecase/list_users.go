package usecase

import (
	"context"

	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

type ListUsersResult struct {
	Users []domain.User
	Total int
}

type ListUsersUseCase struct {
	repo domain.UsersRepository
}

func NewListUsersUseCase(repo domain.UsersRepository) *ListUsersUseCase {
	return &ListUsersUseCase{repo: repo}
}

func (uc *ListUsersUseCase) ListUsers(ctx context.Context, limit, offset int) (*ListUsersResult, error) {
	users, total, err := uc.repo.ListUsers(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	return &ListUsersResult{Users: users, Total: total}, nil
}
