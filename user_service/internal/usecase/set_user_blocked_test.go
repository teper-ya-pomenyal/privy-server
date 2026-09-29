package usecase

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

// Подмены для теста: репозиторий в памяти и хранилище сессий, запоминающее отзывы.
type fakeUsersRepo struct {
	users map[uuid.UUID]*domain.User
}

func (f *fakeUsersRepo) AddUser(ctx context.Context, user *domain.User) error { return nil }
func (f *fakeUsersRepo) GetUserByUserName(ctx context.Context, userName string) (*domain.User, error) {
	return nil, domain.ErrUserNotFound
}
func (f *fakeUsersRepo) GetUserByID(ctx context.Context, userUUID uuid.UUID) (*domain.User, error) {
	u, ok := f.users[userUUID]
	if !ok {
		return nil, domain.ErrUserNotFound
	}
	return u, nil
}
func (f *fakeUsersRepo) UserAlreadyExists(ctx context.Context, userName string) (bool, error) {
	return false, nil
}
func (f *fakeUsersRepo) AnyUserExists(ctx context.Context) (bool, error) {
	return len(f.users) > 0, nil
}
func (f *fakeUsersRepo) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	return nil, 0, nil
}
func (f *fakeUsersRepo) SetUserBlocked(ctx context.Context, userUUID uuid.UUID, blocked bool) error {
	f.users[userUUID].Blocked = blocked
	return nil
}
func (f *fakeUsersRepo) Ping(ctx context.Context) error { return nil }

type fakeSessions struct {
	revoked map[uuid.UUID]int
}

func (f *fakeSessions) Save(ctx context.Context, refreshToken string, userUUID uuid.UUID) error {
	return nil
}
func (f *fakeSessions) Refresh(ctx context.Context, oldToken, newToken string) error { return nil }
func (f *fakeSessions) Get(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	return uuid.UUID{}, domain.ErrRefreshTokenNotFound
}
func (f *fakeSessions) Delete(ctx context.Context, refreshToken string) error { return nil }
func (f *fakeSessions) ListSessions(ctx context.Context, userUUID uuid.UUID) ([]domain.SessionInfo, error) {
	return nil, nil
}
func (f *fakeSessions) RevokeSessions(ctx context.Context, userUUID uuid.UUID, keepSessionID string, remove []string) error {
	f.revoked[userUUID]++
	return nil
}
func (f *fakeSessions) Ping(ctx context.Context) error { return nil }

func newBlockedTestDeps(users ...*domain.User) (*SetUserBlockedUseCase, *fakeUsersRepo, *fakeSessions) {
	repo := &fakeUsersRepo{users: map[uuid.UUID]*domain.User{}}
	for _, u := range users {
		repo.users[u.UserUUID] = u
	}
	sessions := &fakeSessions{revoked: map[uuid.UUID]int{}}
	return NewSetUserBlockedUseCase(repo, sessions), repo, sessions
}

func TestSetBlockedBlocksUserAndRevokesSessions(t *testing.T) {
	regular := &domain.User{UserUUID: uuid.New(), UserName: "ivan", Role: domain.RoleUser}
	uc, repo, sessions := newBlockedTestDeps(regular)

	if err := uc.SetBlocked(context.Background(), regular.UserUUID, true); err != nil {
		t.Fatalf("SetBlocked: %v", err)
	}
	if !repo.users[regular.UserUUID].Blocked {
		t.Fatal("user must be blocked")
	}
	if sessions.revoked[regular.UserUUID] != 1 {
		t.Fatalf("blocking must revoke sessions, got %d", sessions.revoked[regular.UserUUID])
	}

	if err := uc.SetBlocked(context.Background(), regular.UserUUID, false); err != nil {
		t.Fatalf("unblock: %v", err)
	}
	if repo.users[regular.UserUUID].Blocked {
		t.Fatal("user must be unblocked")
	}
	if sessions.revoked[regular.UserUUID] != 1 {
		t.Fatal("unblock must not revoke sessions")
	}
}

func TestSetBlockedRefusesOwner(t *testing.T) {
	owner := &domain.User{UserUUID: uuid.New(), UserName: "owner", Role: domain.RoleOwner}
	uc, repo, sessions := newBlockedTestDeps(owner)

	err := uc.SetBlocked(context.Background(), owner.UserUUID, true)
	if err != domain.ErrCannotBlockOwner {
		t.Fatalf("expected ErrCannotBlockOwner, got %v", err)
	}
	if repo.users[owner.UserUUID].Blocked {
		t.Fatal("owner must not be blocked")
	}
	if len(sessions.revoked) != 0 {
		t.Fatal("no sessions may be revoked when the guard fires")
	}

	// разблокировка (снятие уже стоящей блокировки) не запрещена
	if err := uc.SetBlocked(context.Background(), owner.UserUUID, false); err != nil {
		t.Fatalf("unblock owner: %v", err)
	}
}

func TestSetBlockedUnknownUser(t *testing.T) {
	uc, _, _ := newBlockedTestDeps()
	if err := uc.SetBlocked(context.Background(), uuid.New(), true); err != domain.ErrUserNotFound {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}
