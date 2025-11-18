package redislock

import (
	"context"
	"time"

	rediscli "server/common/pkg/redis"
	"server/common/utils"
)

// 加锁
func AcquireLock(ctx context.Context, key string, ttl time.Duration) (string, bool, error) {
	token := utils.NewULID()

	ok, err := rediscli.Rds.SetNX(ctx, key, token, ttl).Result()
	if err != nil {
		return "", false, err
	}
	return token, ok, nil
}

// 解锁
func ReleaseLock(ctx context.Context, key, token string) error {
	luaScript := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("del", KEYS[1])
        else
            return 0
        end
    `
	_, err := rediscli.Rds.Eval(ctx, luaScript, []string{key}, token).Result()
	return err
}

// 锁续租
func RefreshLock(ctx context.Context, key, token string, ttl time.Duration) (bool, error) {
	luaScript := `
        if redis.call("get", KEYS[1]) == ARGV[1] then
            return redis.call("pexpire", KEYS[1], ARGV[2])
        else
            return 0
        end
    `
	result, err := rediscli.Rds.Eval(ctx, luaScript, []string{key}, token, int64(ttl/time.Millisecond)).Result()
	if err != nil {
		return false, err
	}
	return result.(int64) == 1, nil
}
