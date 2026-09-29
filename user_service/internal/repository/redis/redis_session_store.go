package redis

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/config"
	"github.com/teper-ya-pomenyal/privy_stream/user_service/internal/domain"
)

type RedisSessionStore struct {
	ttlRefresh time.Duration
	conn       *redis.Client
}

func NewRedisSessionStore(cacheDB *config.UserDBCacheConfig, ttlRefresh time.Duration) (*RedisSessionStore, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cacheDB.DBAddress,
		Password:     cacheDB.DBPassword,
		DB:           cacheDB.DBNumber,
		PoolSize:     cacheDB.PoolSize,
		MinIdleConns: cacheDB.MinIdleConns,
		MaxRetries:   cacheDB.MaxRetries,
		DialTimeout:  cacheDB.DialTimeout,
		ReadTimeout:  cacheDB.ReadTimeout,
		WriteTimeout: cacheDB.WriteTimeout,
	})

	ctx := context.Background()
	pong, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}
	log.Println(pong)
	return &RedisSessionStore{
		ttlRefresh: ttlRefresh,
		conn:       rdb,
	}, nil
}

func (r *RedisSessionStore) Save(ctx context.Context, refreshToken string, userUUID uuid.UUID) error {
	zsetKey := userUUID.String()

	pipe := r.conn.TxPipeline()
	pipe.Set(ctx, refreshToken, userUUID.String(), r.ttlRefresh)
	pipe.ZAdd(ctx, zsetKey, redis.Z{
		Member: refreshToken,
		Score:  float64(time.Now().Unix()),
	})
	pipe.Expire(ctx, zsetKey, r.ttlRefresh)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return r.evictIfNeeded(ctx, zsetKey)
}

// Check session
func (r *RedisSessionStore) Get(ctx context.Context, refreshToken string) (uuid.UUID, error) {
	val, err := r.conn.Get(ctx, refreshToken).Result()
	if err == redis.Nil {
		return uuid.UUID{}, domain.ErrRefreshTokenNotFound
	} else if err != nil {
		return uuid.UUID{}, err
	}
	userUUID, err := uuid.Parse(val)
	if err != nil {
		return uuid.UUID{}, err
	}
	return userUUID, nil
}

// Delete session
func (r *RedisSessionStore) Delete(ctx context.Context, refreshToken string) error {
	userUUID, err := r.conn.Get(ctx, refreshToken).Result()
	if err == redis.Nil {
		return domain.ErrRefreshTokenNotFound
	} else if err != nil {
		return err
	}

	pipe := r.conn.TxPipeline()
	pipe.Del(ctx, refreshToken)
	pipe.ZRem(ctx, userUUID, refreshToken)
	if _, err = pipe.Exec(ctx); err != nil {
		return err
	}
	return nil
}

// Refresh sessions
func (r *RedisSessionStore) Refresh(ctx context.Context, oldToken, newToken string) error {
	zsetKey, err := r.conn.Get(ctx, oldToken).Result()
	if err == redis.Nil {
		return domain.ErrRefreshTokenNotFound
	} else if err != nil {
		return err
	}

	pipe := r.conn.TxPipeline()
	pipe.Del(ctx, oldToken)
	pipe.ZRem(ctx, zsetKey, oldToken)
	pipe.Set(ctx, newToken, zsetKey, r.ttlRefresh)
	pipe.ZAdd(ctx, zsetKey, redis.Z{
		Member: newToken,
		Score:  float64(time.Now().Unix()),
	})
	pipe.Expire(ctx, zsetKey, r.ttlRefresh)
	if _, err := pipe.Exec(ctx); err != nil {
		return err
	}

	return r.evictIfNeeded(ctx, zsetKey)
}

func (r *RedisSessionStore) evictIfNeeded(ctx context.Context, zsetKey string) error {
	const maxSessions = 7

	count, err := r.conn.ZCard(ctx, zsetKey).Result()
	if err != nil {
		return err
	}
	if count <= maxSessions {
		return nil
	}

	oldToken, err := r.conn.ZRange(ctx, zsetKey, 0, 0).Result()
	if err != nil {
		return err
	}
	if len(oldToken) == 0 {
		return nil
	}

	evictPipe := r.conn.TxPipeline()
	evictPipe.Del(ctx, oldToken[0])
	evictPipe.ZRem(ctx, zsetKey, oldToken[0])
	_, err = evictPipe.Exec(ctx)
	return err
}

// ListSessions возвращает активные refresh-сессии пользователя: score ZSET —
// время последней ротации токена, TTL ключа — срок до истечения. Токены, уже
// истёкшие между ZRange и PTTL, пропускаются.
func (r *RedisSessionStore) ListSessions(ctx context.Context, userUUID uuid.UUID) ([]domain.SessionInfo, error) {
	members, err := r.conn.ZRangeWithScores(ctx, userUUID.String(), 0, -1).Result()
	if err != nil {
		return nil, err
	}

	now := time.Now()
	sessions := make([]domain.SessionInfo, 0, len(members))
	for _, m := range members {
		token, ok := m.Member.(string)
		if !ok {
			continue
		}
		ttl, err := r.conn.PTTL(ctx, token).Result()
		if err != nil || ttl <= 0 {
			continue
		}
		sessions = append(sessions, domain.SessionInfo{
			SessionID: sessionID(token),
			CreatedAt: time.Unix(int64(m.Score), 0),
			ExpiresAt: now.Add(ttl),
		})
	}
	return sessions, nil
}

// RevokeSessions удаляет сессии по хэшам: перечисленные в remove либо,
// если список пуст, все сессии пользователя, кроме keepSessionID
// (пустой keepSessionID при пустом remove — все сессии).
func (r *RedisSessionStore) RevokeSessions(ctx context.Context, userUUID uuid.UUID, keepSessionID string, remove []string) error {
	members, err := r.conn.ZRange(ctx, userUUID.String(), 0, -1).Result()
	if err != nil {
		return err
	}

	removeSet := make(map[string]struct{}, len(remove))
	for _, id := range remove {
		removeSet[id] = struct{}{}
	}

	victims := selectSessionVictims(members, keepSessionID, removeSet)
	if len(victims) == 0 {
		return nil
	}
	pipe := r.conn.TxPipeline()
	for _, token := range victims {
		pipe.Del(ctx, token)
		pipe.ZRem(ctx, userUUID.String(), token)
	}
	_, err = pipe.Exec(ctx)
	return err
}

func sessionID(refreshToken string) string {
	sum := sha256.Sum256([]byte(refreshToken))
	return hex.EncodeToString(sum[:])
}

// selectSessionVictims выбирает токены к удалению: если removeSet непуст —
// только перечисленные хэши, иначе все, кроме keepSessionID
// (пустой keepSessionID — все токены).
func selectSessionVictims(members []string, keepSessionID string, removeSet map[string]struct{}) []string {
	var victims []string
	for _, token := range members {
		id := sessionID(token)
		if len(removeSet) > 0 {
			if _, ok := removeSet[id]; !ok {
				continue
			}
		} else if keepSessionID != "" && id == keepSessionID {
			continue
		}
		victims = append(victims, token)
	}
	return victims
}

// Ping — проверка доступности redis для health.
func (r *RedisSessionStore) Ping(ctx context.Context) error {
	return r.conn.Ping(ctx).Err()
}
