package jwtmanager

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AccessClaims struct {
	jwt.RegisteredClaims
	BirthDate time.Time `json:"birth_date"`
	Role      string    `json:"role,omitempty"`
}

type StreamClaims struct {
	jwt.RegisteredClaims
	Scope     string    `json:"scope"`
	BirthDate time.Time `json:"birth_date"`
}
