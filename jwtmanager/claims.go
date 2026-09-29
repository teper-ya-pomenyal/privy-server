package jwtmanager

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	jwt.RegisteredClaims
	BirthDate time.Time `json:"birth_date"`
	// Роль на узле: owner получает право менять каталог (см. gateway RequireOwner).
	Role string `json:"role,omitempty"`
}
