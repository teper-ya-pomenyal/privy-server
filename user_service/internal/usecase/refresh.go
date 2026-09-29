package usecase

import (
	"context"

	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

type RefreshResult struct {
	AccessToken  string
	RefreshToken string
}

type RefreshUseCase struct {
	repo           domain.UsersRepository
	tokenManager   TokenManager
	sessionManager SessionStore
}

func NewRefreshUseCase(repo domain.UsersRepository, tokenManager TokenManager, sessionManager SessionStore) *RefreshUseCase {
	return &RefreshUseCase{
		repo:           repo,
		tokenManager:   tokenManager,
		sessionManager: sessionManager,
	}
}

func (r *RefreshUseCase) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	// validation refresh token
	userUUID, err := r.sessionManager.Get(ctx, refreshToken)
	if err == domain.ErrRefreshTokenNotFound {
		return nil, domain.ErrRefreshTokenNotFound
	}
	if err != nil {
		return &RefreshResult{}, err
	}

	user, err := r.repo.GetUserByID(ctx, userUUID)
	if err != nil {
		return &RefreshResult{}, err
	}
	// Заблокированный аккаунт не обновляет пару токенов: access истечёт сам,
	// refresh отозван при блокировке.
	if user.Blocked {
		return &RefreshResult{}, domain.ErrUserBlocked
	}

	// make response
	newRefreshToken, err := r.tokenManager.NewRefreshToken()
	if err != nil {
		return &RefreshResult{}, err
	}

	newAccessToken, err := r.tokenManager.NewAccessToken(userUUID, user.BirthDate, user.Role)
	if err != nil {
		return &RefreshResult{}, err
	}

	result := &RefreshResult{
		RefreshToken: newRefreshToken,
		AccessToken:  newAccessToken,
	}
	err = r.sessionManager.Refresh(ctx, refreshToken, newRefreshToken)
	if err != nil {
		return &RefreshResult{}, err
	}
	return result, nil
}
