package clients

import (
	"context"
	"time"

	"github.com/google/uuid"
	userv1 "github.com/teper-ya-pomenyal/privy_stream/proto/user/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type LoginResult struct {
	UserUUID     string    `json:"user_uuid"`
	RefreshToken string    `json:"refresh_token"`
	AccessToken  string    `json:"access_token"`
	BirthDate    time.Time `json:"birth_date"`
}

type RefreshResult struct {
	RefreshToken string `json:"refresh_token"`
	AccessToken  string `json:"access_token"`
}

// SessionEntry — активная refresh-сессия; session_id = SHA-256 refresh-токена,
// сам токен из user_service не выходит.
type SessionEntry struct {
	SessionID string `json:"session_id"`
	CreatedAt int64  `json:"created_at"` // unix seconds — последняя ротация
	ExpiresAt int64  `json:"expires_at"` // unix seconds
}

type UserEntry struct {
	UserUUID  string `json:"user_uuid"`
	UserName  string `json:"user_name"`
	Role      string `json:"role"`
	Blocked   bool   `json:"blocked"`
	BirthDate string `json:"birth_date"` // RFC3339
	CreatedAt string `json:"created_at"` // RFC3339
}

type UsersPage struct {
	Users []UserEntry `json:"users"`
	Total int32       `json:"total"`
}

type UserClient struct {
	conn       *grpc.ClientConn
	grpcClient userv1.UserServiceClient
}

func NewUserClient(address string) (*UserClient, error) {
	conn, err := grpc.NewClient(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &UserClient{conn: conn, grpcClient: userv1.NewUserServiceClient(conn)}, nil
}

func (u *UserClient) Login(ctx context.Context, userName, password string) (*LoginResult, error) {
	req := &userv1.LoginRequest{
		UserName: userName,
		Password: password,
	}
	res, err := u.grpcClient.Login(ctx, req)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		UserUUID:     res.UserUuid,
		RefreshToken: res.RefreshToken,
		AccessToken:  res.AccessToken,
		BirthDate:    res.BirthDate.AsTime(),
	}, nil
}

func (u *UserClient) Register(ctx context.Context, userName, password string, birthDate time.Time) (*LoginResult, error) {
	req := &userv1.RegisterRequest{
		UserName:  userName,
		Password:  password,
		BirthDate: timestamppb.New(birthDate),
	}

	res, err := u.grpcClient.Register(ctx, req)
	if err != nil {
		return nil, err
	}
	return &LoginResult{
		UserUUID:     res.UserUuid,
		RefreshToken: res.RefreshToken,
		AccessToken:  res.AccessToken,
		BirthDate:    res.BirthDate.AsTime(),
	}, nil
}

func (u *UserClient) Refresh(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	req := &userv1.RefreshRequest{RefreshToken: refreshToken}

	res, err := u.grpcClient.Refresh(ctx, req)
	if err != nil {
		return nil, err
	}
	return &RefreshResult{
		RefreshToken: res.RefreshToken,
		AccessToken:  res.AccessToken,
	}, nil
}

func (u *UserClient) Logout(ctx context.Context, refreshToken string) error {
	req := &userv1.LogoutRequest{RefreshToken: refreshToken}

	_, err := u.grpcClient.Logout(ctx, req)
	if err != nil {
		return err
	}
	return nil
}

func (u *UserClient) ListSessions(ctx context.Context, userUUID string) ([]SessionEntry, error) {
	res, err := u.grpcClient.ListSessions(ctx, &userv1.ListSessionsRequest{UserUuid: userUUID})
	if err != nil {
		return nil, err
	}
	sessions := make([]SessionEntry, 0, len(res.Sessions))
	for _, s := range res.Sessions {
		sessions = append(sessions, SessionEntry{SessionID: s.SessionId, CreatedAt: s.CreatedAt, ExpiresAt: s.ExpiresAt})
	}
	return sessions, nil
}

func (u *UserClient) RevokeSessions(ctx context.Context, userUUID, keepSessionID string, sessionIDs []string) error {
	_, err := u.grpcClient.RevokeSessions(ctx, &userv1.RevokeSessionsRequest{
		UserUuid:      userUUID,
		KeepSessionId: keepSessionID,
		SessionIds:    sessionIDs,
	})
	return err
}

func (u *UserClient) ListUsers(ctx context.Context, limit, offset int32) (*UsersPage, error) {
	res, err := u.grpcClient.ListUsers(ctx, &userv1.ListUsersRequest{Limit: limit, Offset: offset})
	if err != nil {
		return nil, err
	}
	users := make([]UserEntry, 0, len(res.Users))
	for _, usr := range res.Users {
		users = append(users, UserEntry{
			UserUUID:  usr.UserUuid,
			UserName:  usr.UserName,
			Role:      usr.Role,
			Blocked:   usr.Blocked,
			BirthDate: usr.BirthDate.AsTime().Format(time.RFC3339),
			CreatedAt: usr.CreatedAt.AsTime().Format(time.RFC3339),
		})
	}
	return &UsersPage{Users: users, Total: res.Total}, nil
}

func (u *UserClient) SetUserBlocked(ctx context.Context, userUUID string, blocked bool) error {
	_, err := u.grpcClient.SetUserBlocked(ctx, &userv1.SetUserBlockedRequest{UserUuid: userUUID, Blocked: blocked})
	return err
}

// Health возвращает доступность postgres и redis (сессии) в user_service.
func (u *UserClient) Health(ctx context.Context) (postgres, redis bool, err error) {
	res, err := u.grpcClient.Health(ctx, &userv1.HealthRequest{})
	if err != nil {
		return false, false, err
	}
	return res.Postgres, res.Redis, nil
}

func (u *UserClient) CreateStreamToken(ctx context.Context, userUUID uuid.UUID, birthDate time.Time) (string, error) {
	req := userv1.CreateStreamTokenRequest{
		UserUuid:  userUUID.String(),
		BirthDate: timestamppb.New(birthDate),
	}
	res, err := u.grpcClient.CreateStreamToken(ctx, &req)
	if err != nil {
		return "", err
	}
	return res.StreamToken, nil
}
