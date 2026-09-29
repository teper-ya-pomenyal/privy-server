package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/config"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

type UsersPostgresRepository struct {
	conn *pgxpool.Pool
}

func NewUsersPostgresRepository(cfg *config.UserDBConfig) (*UsersPostgresRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cfgDB, err := pgxpool.ParseConfig(cfg.DSN)
	if err != nil {
		return nil, err
	}
	cfgDB.MaxConns = cfg.MaxConns
	cfgDB.MinConns = cfg.MinConns
	cfgDB.MaxConnLifetime = cfg.MaxConnLifetime
	cfgDB.MaxConnIdleTime = cfg.MaxConnIdleTime
	cfgDB.HealthCheckPeriod = cfg.HealthCheckPeriod
	cfgDB.ConnConfig.ConnectTimeout = cfg.ConnectTimeout
	conn, err := pgxpool.NewWithConfig(ctx, cfgDB)
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(ctx); err != nil {
		conn.Close()
		return nil, err
	}

	return &UsersPostgresRepository{conn: conn}, nil
}

func (p *UsersPostgresRepository) Close() {
	p.conn.Close()
}

func (p *UsersPostgresRepository) AddUser(ctx context.Context, user *domain.User) error {
	_, err := p.conn.Exec(ctx,
		"INSERT INTO users (uuid, user_name, password_hash, birth_date, role, blocked, created_at) VALUES($1, $2, $3, $4, $5, $6, $7)",
		user.UserUUID, user.UserName, user.PasswordHash, user.BirthDate, user.Role, user.Blocked, user.CreatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrUserAlreadyExists
		}
		return err
	}
	return nil

}

const userColumns = "uuid, user_name, password_hash, birth_date, role, blocked, created_at"

func scanUser(row pgx.Row) (*domain.User, error) {
	u := &domain.User{}
	err := row.Scan(&u.UserUUID, &u.UserName, &u.PasswordHash, &u.BirthDate, &u.Role, &u.Blocked, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (p *UsersPostgresRepository) GetUserByUserName(ctx context.Context, userName string) (*domain.User, error) {
	return scanUser(p.conn.QueryRow(ctx,
		"SELECT "+userColumns+" FROM users WHERE user_name = $1",
		userName,
	))
}

func (p *UsersPostgresRepository) GetUserByID(ctx context.Context, userUUID uuid.UUID) (*domain.User, error) {
	return scanUser(p.conn.QueryRow(ctx,
		"SELECT "+userColumns+" FROM users WHERE uuid = $1",
		userUUID,
	))
}

func (p *UsersPostgresRepository) UserAlreadyExists(ctx context.Context, userName string) (bool, error) {
	var exists bool
	err := p.conn.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE user_name = $1)",
		userName,
	).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil

}

func (p *UsersPostgresRepository) AnyUserExists(ctx context.Context) (bool, error) {
	var exists bool
	err := p.conn.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM users)").Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

func (p *UsersPostgresRepository) ListUsers(ctx context.Context, limit, offset int) ([]domain.User, int, error) {
	var total int
	if err := p.conn.QueryRow(ctx, "SELECT COUNT(*) FROM users").Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := p.conn.Query(ctx,
		"SELECT "+userColumns+" FROM users ORDER BY created_at, uuid LIMIT $1 OFFSET $2",
		limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	users := []domain.User{}
	for rows.Next() {
		u := &domain.User{}
		if err := rows.Scan(&u.UserUUID, &u.UserName, &u.PasswordHash, &u.BirthDate, &u.Role, &u.Blocked, &u.CreatedAt); err != nil {
			return nil, 0, err
		}
		users = append(users, *u)
	}
	if err = rows.Err(); err != nil {
		return nil, 0, err
	}
	return users, total, nil
}

func (p *UsersPostgresRepository) SetUserBlocked(ctx context.Context, userUUID uuid.UUID, blocked bool) error {
	res, err := p.conn.Exec(ctx,
		"UPDATE users SET blocked = $2 WHERE uuid = $1",
		userUUID, blocked,
	)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

func (p *UsersPostgresRepository) Ping(ctx context.Context) error {
	return p.conn.Ping(ctx)
}
