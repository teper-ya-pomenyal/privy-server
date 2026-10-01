package grpc

import (
	userv1 "github.com/teper-ya-pomenyal/privy_stream/proto/user/v1"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/usecase"
)

type UserGRPCHandler struct {
	userv1.UnimplementedUserServiceServer
	loginUseCase             *usecase.LoginUseCase
	registerUseCase          *usecase.RegisterUseCase
	refreshUseCase           *usecase.RefreshUseCase
	logoutUseCase            *usecase.LogoutUseCase
	listSessionsUseCase      *usecase.ListSessionsUseCase
	revokeSessionsUseCase    *usecase.RevokeSessionsUseCase
	listUsersUseCase         *usecase.ListUsersUseCase
	setUserBlockedUseCase    *usecase.SetUserBlockedUseCase
	healthUseCase            *usecase.HealthUseCase
	createStreamTokenUseCase *usecase.CreateStreamTokenUseCase
}

func NewUserGRPCHandler(
	loginUseCase *usecase.LoginUseCase,
	registerUseCase *usecase.RegisterUseCase,
	refreshUseCase *usecase.RefreshUseCase,
	logoutUseCase *usecase.LogoutUseCase,
	listSessionsUseCase *usecase.ListSessionsUseCase,
	revokeSessionsUseCase *usecase.RevokeSessionsUseCase,
	listUsersUseCase *usecase.ListUsersUseCase,
	setUserBlockedUseCase *usecase.SetUserBlockedUseCase,
	healthUseCase *usecase.HealthUseCase,
	streamToken *usecase.CreateStreamTokenUseCase,
) *UserGRPCHandler {
	return &UserGRPCHandler{
		loginUseCase:             loginUseCase,
		registerUseCase:          registerUseCase,
		refreshUseCase:           refreshUseCase,
		logoutUseCase:            logoutUseCase,
		listSessionsUseCase:      listSessionsUseCase,
		revokeSessionsUseCase:    revokeSessionsUseCase,
		listUsersUseCase:         listUsersUseCase,
		setUserBlockedUseCase:    setUserBlockedUseCase,
		healthUseCase:            healthUseCase,
		createStreamTokenUseCase: streamToken,
	}
}
