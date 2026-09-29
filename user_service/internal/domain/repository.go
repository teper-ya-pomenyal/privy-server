package domain

import (
	"context"

	"github.com/google/uuid"
)

type UsersRepository interface {
	AddUser(ctx context.Context, user *User) error
	GetUserByUserName(ctx context.Context, userName string) (*User, error)
	GetUserByID(ctx context.Context, userUUID uuid.UUID) (*User, error)
	UserAlreadyExists(ctx context.Context, userName string) (bool, error)
	// Есть ли в таблице хоть один аккаунт: первый на узле становится владельцем.
	AnyUserExists(ctx context.Context) (bool, error)
	// Admin: список аккаунтов с общим счётчиком и смена блокировки.
	ListUsers(ctx context.Context, limit, offset int) ([]User, int, error)
	SetUserBlocked(ctx context.Context, userUUID uuid.UUID, blocked bool) error
	Ping(ctx context.Context) error
}
