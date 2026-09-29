package usecase

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

type RegisterUseCase struct {
	repo           domain.UsersRepository
	tokenManager   TokenManager
	sessionManager SessionStore
}

func NewRegisterUseCase(repo domain.UsersRepository, tokenManager TokenManager, sessionManager SessionStore) *RegisterUseCase {
	return &RegisterUseCase{
		repo:           repo,
		tokenManager:   tokenManager,
		sessionManager: sessionManager,
	}
}

func (r *RegisterUseCase) Register(ctx context.Context, userName, password string, birthDate time.Time) (*LoginResult, error) {
	// validate param
	err := utils.ValidateUsername(userName)
	if err != nil {
		return &LoginResult{}, err
	}
	err = utils.ValidatePassword(password)
	if err != nil {
		return &LoginResult{}, err
	}
	err = utils.ValidateIncomingDate(birthDate)
	if err != nil {
		return &LoginResult{}, err
	}

	ok, err := r.repo.UserAlreadyExists(ctx, userName)
	if err != nil {
		return &LoginResult{}, err
	}
	if ok {
		return &LoginResult{}, domain.ErrUserAlreadyExists
	}
	// Роль: самый первый аккаунт узла — его владелец. Гонки двух первых
	// регистраций игнорируем: узел поднимает один человек.
	role := domain.RoleUser
	any, err := r.repo.AnyUserExists(ctx)
	if err != nil {
		return &LoginResult{}, err
	}
	if !any {
		role = domain.RoleOwner
	}
	hashPassword, err := bcrypt.GenerateFromPassword([]byte(password), 10) //установил дефолтное значение для ясности - какое колличество рацндов хэширования.
	if err != nil {
		return &LoginResult{}, err
	}

	newUser := &domain.User{
		UserUUID:     uuid.New(),
		UserName:     userName,
		PasswordHash: string(hashPassword),
		BirthDate:    birthDate,
		Role:         role,
		CreatedAt:    time.Now(),
	}

	err = r.repo.AddUser(ctx, newUser)
	if err != nil {
		return &LoginResult{}, err
	}

	//make session

	accessToken, err := r.tokenManager.NewAccessToken(newUser.UserUUID, newUser.BirthDate, newUser.Role)
	if err != nil {
		return &LoginResult{}, err
	}
	refreshToken, err := r.tokenManager.NewRefreshToken()
	if err != nil {
		return &LoginResult{}, err
	}
	resultLogin := &LoginResult{
		UUID:         newUser.UserUUID.String(),
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		BirthDate:    newUser.BirthDate,
		CreatedAt:    newUser.CreatedAt,
	}
	err = r.sessionManager.Save(ctx, refreshToken, newUser.UserUUID)
	if err != nil {
		return &LoginResult{}, err
	}
	return resultLogin, nil
}
