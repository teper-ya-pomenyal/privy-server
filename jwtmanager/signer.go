package jwtmanager

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Manager struct {
	privateKey *rsa.PrivateKey
	accessTTL  time.Duration
	streamTTL  time.Duration
}

func NewManager(privateKey *rsa.PrivateKey, accessTTL time.Duration, streamTTL time.Duration) *Manager {
	return &Manager{privateKey: privateKey, accessTTL: accessTTL, streamTTL: streamTTL}
}

func (m *Manager) NewAccessToken(userUUID uuid.UUID, birthDate time.Time, role string) (string, error) {
	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.accessTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		BirthDate: birthDate,
		Role:      role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

func (m *Manager) NewStreamToken(userUUID uuid.UUID, birthDate time.Time) (string, error) {
	claims := StreamClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userUUID.String(),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(m.streamTTL)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		BirthDate: birthDate,
		Scope:     "stream",
	}
	token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	return token.SignedString(m.privateKey)
}

func (m *Manager) NewRefreshToken() (string, error) {
	randBytes := make([]byte, 32)
	if _, err := rand.Read(randBytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(randBytes), nil
}
