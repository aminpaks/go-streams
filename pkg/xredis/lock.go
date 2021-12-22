package xredis

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/go-redis/redis/v8"
)

func lock(c *redis.Client, key string) (release func() error) {
	ctx := context.Background()
	lockKey := "LOCK::" + key
	callerInfo := ""
	if _, file, line, ok := runtime.Caller(1); ok {
		callerInfo = fmt.Sprintf("%s:%d", file, line)
	}

	for {
		if b, err := c.SetNX(ctx, lockKey, "LOCK", time.Second*10).Result(); err != nil || !b {
			time.Sleep(time.Millisecond * 1)
			continue
		}
		break
	}

	return func() error {
		if _, err := c.Del(ctx, lockKey).Result(); err != nil {
			return fmt.Errorf("failed to remove lock '%s' from %s", key, callerInfo)
		}
		return nil
	}
}
