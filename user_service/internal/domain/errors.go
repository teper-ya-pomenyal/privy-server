package domain

import "errors"

var (
	ErrUserNotFound         = errors.New("user not found")
	ErrInvalidCredentials   = errors.New("invalid credentials")
	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrRefreshTokenNotFound = errors.New("refresh token not found")
	ErrInvalidDate          = errors.New("invalid date")
	ErrInvalidUUID          = errors.New("invalid uuid")
	ErrUserBlocked          = errors.New("user is blocked")
)

var (
	ErrPasswordTooShort          = errors.New("password is too short")
	ErrPasswordTooLong           = errors.New("password is too long")
	ErrInvalidCharacters         = errors.New("invalid characters")
	ErrMissingRequiredCharacters = errors.New("missing required characters.")
	ErrLoginTooShort             = errors.New("login is too short")
	ErrLoginTooLong              = errors.New("login is too long")
	ErrInvalidRefreshToken       = errors.New("invalid refresh token")
	ErrLoginRequired             = errors.New("login is required")
	ErrPasswordRequired          = errors.New("password is required")
	ErrBirthDateRequired         = errors.New("birth date is required")
)
