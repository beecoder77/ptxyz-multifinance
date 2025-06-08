package redis

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisClient interface defines the Redis operations needed by our application
type RedisClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) *redis.BoolCmd
	Del(ctx context.Context, keys ...string) *redis.IntCmd
	Eval(ctx context.Context, script string, keys []string, args ...interface{}) *redis.Cmd
	Get(ctx context.Context, key string) *redis.StringCmd
}

// Ensure redis.Client implements RedisClient interface
var _ RedisClient = (*redis.Client)(nil)
