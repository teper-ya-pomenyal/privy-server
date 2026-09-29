package grpc

import (
	"errors"

	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// errors mapping
func mapDomainError(err error) error {
	switch {
	case errors.Is(err, domain.ErrInvalidCredentials):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrUserNotFound):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrUserAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, domain.ErrRefreshTokenNotFound):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrInvalidRefreshToken):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, domain.ErrUserBlocked):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, domain.ErrInvalidUUID):
		return status.Error(codes.InvalidArgument, err.Error())

	// login validation
	case errors.Is(err, domain.ErrLoginRequired):
		return fieldViolation(err, "user_name", "login_required")
	case errors.Is(err, domain.ErrLoginTooShort):
		return fieldViolation(err, "user_name", "login_too_short")
	case errors.Is(err, domain.ErrLoginTooLong):
		return fieldViolation(err, "user_name", "login_too_long")
	case errors.Is(err, domain.ErrInvalidCharacters):
		return fieldViolation(err, "user_name", "login_invalid_characters")

	// password validation
	case errors.Is(err, domain.ErrPasswordRequired):
		return fieldViolation(err, "password", "password_required")
	case errors.Is(err, domain.ErrPasswordTooShort):
		return fieldViolation(err, "password", "password_too_short")
	case errors.Is(err, domain.ErrPasswordTooLong):
		return fieldViolation(err, "password", "password_too_long")
	case errors.Is(err, domain.ErrMissingRequiredCharacters):
		return fieldViolation(err, "password", "password_missing_special")

	// birth date validation
	case errors.Is(err, domain.ErrBirthDateRequired):
		return fieldViolation(err, "birth_date", "birth_date_required")
	case errors.Is(err, domain.ErrInvalidDate):
		return fieldViolation(err, "birth_date", "birth_date_out_of_range")
	default:
		return status.Error(codes.Internal, "internal error")

	}
}

func fieldViolation(err error, field, reason string) error {
	st, detailsErr := status.New(codes.InvalidArgument, err.Error()).WithDetails(&errdetails.BadRequest{
		FieldViolations: []*errdetails.BadRequest_FieldViolation{
			{Field: field, Reason: reason, Description: err.Error()},
		},
	})
	if detailsErr != nil {
		return status.Error(codes.InvalidArgument, err.Error())
	}
	return st.Err()
}
