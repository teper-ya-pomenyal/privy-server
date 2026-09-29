package domain

import "time"

// SessionInfo — активная refresh-сессия. SessionID — это SHA-256 refresh-токена:
// сам токен не покидает user_service, а по хэшу сессию можно отозвать.
type SessionInfo struct {
	SessionID string
	CreatedAt time.Time // последняя ротация токена
	ExpiresAt time.Time
}
