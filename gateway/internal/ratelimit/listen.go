package ratelimit

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// ListenLimiter не даёт одному пользователю засчитывать прослушивание
// одного и того же трека чаще, чем раз в window.
type ListenLimiter struct {
	rdb    *redis.Client
	window time.Duration
}

func NewListenLimiter(rdb *redis.Client, window time.Duration) *ListenLimiter {
	return &ListenLimiter{rdb: rdb, window: window}
}

func key(userID, trackUUID string) string {
	return "listened:" + userID + ":" + trackUUID
}

// Allow атомарно занимает слот (SET NX EX). false — слот уже занят.
func (l *ListenLimiter) Allow(ctx context.Context, userID, trackUUID string) (bool, error) {
	return l.rdb.SetNX(ctx, key(userID, trackUUID), 1, l.window).Result()
}

// Release освобождает слот, если прослушивание так и не засчитали.
func (l *ListenLimiter) Release(ctx context.Context, userID, trackUUID string) error {
	return l.rdb.Del(ctx, key(userID, trackUUID)).Err()
}
