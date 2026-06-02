package repository

import (
	"context"
	"time"
)

type KeyValueRepository interface {
	Set(ctx context.Context, key, value string)
	SetEX(ctx context.Context, key, value string, ttl time.Duration)
	SetNX(ctx context.Context, key, value string, ttl time.Duration) bool
	Get(ctx context.Context, key string) (string, bool)
	GetDel(ctx context.Context, key string) (string, bool)
	Del(ctx context.Context, key string) int
	Expire(ctx context.Context, key string, seconds int) bool
	TTL(ctx context.Context, key string) int64
	Persist(ctx context.Context, key string) bool
	Keys(ctx context.Context, pattern string) []string
	Exists(ctx context.Context, key string) bool
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, int64, bool)
	AcquireLock(ctx context.Context, key, owner string, ttl time.Duration) bool
	ReleaseLock(ctx context.Context, key, owner string) bool
	Touch(ctx context.Context, key string, ttl time.Duration) bool
	Size(ctx context.Context) int
	StartCleanup(intervalMs int64)
	StopCleanup()
}
