package usecase

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/utils"
)

var dummyHash, _ = bcrypt.GenerateFromPassword([]byte("dummy-password"), 10)

type LoginResult struct {
	UUID         string
	AccessToken  string
	RefreshToken string
	BirthDate    time.Time
	CreatedAt    time.Time
}

type LoginUseCase struct {
	repo           domain.UsersRepository
	tokenManager   TokenManager
	sessionManager SessionStore
}

func NewLoginUseCase(repo domain.UsersRepository, tokenManager TokenManager, sessionManager SessionStore) *LoginUseCase {
	return &LoginUseCase{
		repo:           repo,
		tokenManager:   tokenManager,
		sessionManager: sessionManager,
	}
}

func (l *LoginUseCase) Login(ctx context.Context, userName, password string) (*LoginResult, error) {
	// user search
	err := utils.ValidateUsername(userName)
	if err != nil {
		return nil, err
	}
	err = utils.ValidatePassword(password)
	if err != nil {
		return nil, err
	}

	user, err := l.repo.GetUserByUserName(ctx, userName)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			_ = bcrypt.CompareHashAndPassword(dummyHash, []byte(password))
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	// Блокировка не выдаёт себя через неверный пароль — отвечает явно.
	if user.Blocked {
		return nil, domain.ErrUserBlocked
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	//make session

	accessToken, err := l.tokenManager.NewAccessToken(user.UserUUID, user.BirthDate, user.Role)
	if err != nil {
		return nil, err
	}
	refreshToken, err := l.tokenManager.NewRefreshToken()
	if err != nil {
		return nil, err
	}
	loginResult := &LoginResult{
		UUID:         user.UserUUID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		BirthDate:    user.BirthDate,
		CreatedAt:    user.CreatedAt,
	}
	err = l.sessionManager.Save(ctx, refreshToken, user.UserUUID)
	if err != nil {
		return nil, err
	}
	return loginResult, nil
}
