package redis

import (
	"context"
	"fmt"
	"time"
)

// DistributedLock represents a distributed lock implementation using Redis
type DistributedLock struct {
	client RedisClient
	key    string
	value  string
	ttl    time.Duration
}

// NewDistributedLock creates a new distributed lock instance
func NewDistributedLock(client RedisClient, key string, ttl time.Duration) *DistributedLock {
	return &DistributedLock{
		client: client,
		key:    fmt.Sprintf("lock:%s", key),
		value:  fmt.Sprintf("%d", time.Now().UnixNano()),
		ttl:    ttl,
	}
}

// Lock attempts to acquire the lock
func (dl *DistributedLock) Lock(ctx context.Context) error {
	// Try to set the lock key with NX option (only if it doesn't exist)
	result, err := dl.client.SetNX(ctx, dl.key, dl.value, dl.ttl).Result()
	if err != nil {
		return fmt.Errorf("failed to acquire lock: %v", err)
	}
	if !result {
		return fmt.Errorf("lock already held")
	}
	return nil
}

// Unlock releases the lock
func (dl *DistributedLock) Unlock(ctx context.Context) error {
	// Use Lua script to ensure we only delete our own lock
	script := `
		if redis.call("get", KEYS[1]) == ARGV[1] then
			return redis.call("del", KEYS[1])
		else
			return 0
		end
	`

	result, err := dl.client.Eval(ctx, script, []string{dl.key}, dl.value).Int64()
	if err != nil {
		return fmt.Errorf("failed to release lock: %v", err)
	}
	if result == 0 {
		return fmt.Errorf("lock not held or expired")
	}
	return nil
}

// TryLock attempts to acquire the lock with timeout
func (dl *DistributedLock) TryLock(ctx context.Context, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		err := dl.Lock(ctx)
		if err == nil {
			return nil
		}
		// Wait a bit before retrying
		time.Sleep(100 * time.Millisecond)
	}
	return fmt.Errorf("timeout acquiring lock")
}

// Example usage:
// lock := NewDistributedLock(redisClient, "transaction:123", 30*time.Second)
// if err := lock.TryLock(ctx, 5*time.Second); err != nil {
//     return err
// }
// defer lock.Unlock(ctx)
