package usecase

import (
	"time"

	"github.com/google/uuid"
)

type TokenManager interface {
	NewAccessToken(userUUID uuid.UUID, birthDate time.Time, role string) (string, error)
	NewRefreshToken() (string, error)
	NewStreamToken(userUUID uuid.UUID, birthDate time.Time) (string, error)
}
