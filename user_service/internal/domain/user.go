package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	RoleUser  = "user"
	RoleOwner = "owner"
)

type User struct {
	UserUUID     uuid.UUID `db:"uuid"`
	UserName     string    `db:"user_name"`
	PasswordHash string    `db:"password_hash"`
	BirthDate    time.Time `db:"birth_date"`
	Role         string    `db:"role"`
	Blocked      bool      `db:"blocked"`
	CreatedAt    time.Time `db:"created_at"`
}
