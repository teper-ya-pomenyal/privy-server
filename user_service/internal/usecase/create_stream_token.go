package usecase

import (
	"time"

	"github.com/google/uuid"
)

type CreateStreamTokenUseCase struct {
	tokenManager TokenManager
}

func NewCreateStreamTokenUseCase(tm TokenManager) *CreateStreamTokenUseCase {
	return &CreateStreamTokenUseCase{tokenManager: tm}
}

func (u *CreateStreamTokenUseCase) CreateStreamToken(userUUID uuid.UUID, birthDate time.Time) (string, error) {
	token, err := u.tokenManager.NewStreamToken(userUUID, birthDate)
	if err != nil {
		return "", err
	}
	return token, nil
}
