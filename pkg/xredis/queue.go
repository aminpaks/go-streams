package xredis

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis"
	"github.com/google/uuid"
)

var ctx = context.Background()

type XQueuer func(queueName string, entries ...XQueueEntry) (referenceUris []string, errs map[string]error)

func NewQueueConsumer(depsCtx context.Context, queueName string, count int, consumerFn func(val ...XQueueEntry)) error {
	rdb, err := GetClient(depsCtx)
	if err != nil {
		return err
	}

	if count < 1 {
		count = 1
	}
	if queueName == "" {
		queueName = "randomQueueName" + strconv.Itoa(rand.Int())
	}

	go func() {
		for {
			values := []XQueueEntry{}
			for i := 0; i < count; i++ {
				entry, err := internalPopFromTop(rdb, queueName)
				if err != nil {
					log.Printf("queue %s: %v", queueName, err)
				}
				if entry != nil {
					values = append(values, *entry)
				}
			}
			if len(values) > 0 {
				consumerFn(values...)
			}
		}
	}()

	return nil
}

func GetQueueValue(referenceUri string) (string, error) {
	return "", nil
}

func BuildQueuer(depsCtx context.Context) (XQueuer, error) {
	rdb, err := GetClient(depsCtx)
	if err != nil {
		return nil, err
	}

	return func(queueName string, entries ...XQueueEntry) (referenceUris []string, errs map[string]error) {
		errs = make(map[string]error)
		referenceUris = make([]string, 0)

		for _, entry := range entries {
			referenceUri, err := enqueueEntry(rdb, queueName, entry)
			if err != nil {
				errs[referenceUri] = err
			}
			referenceUris = append(referenceUris, referenceUri)
		}

		return referenceUris, errs
	}, nil
}

func enqueueEntry(rdb *RedisClient, queueName string, entry XQueueEntry) (referenceUri string, err error) {
	if entry.ReferenceUri == "" {
		entry.ReferenceUri = fmt.Sprintf("gid://%s/%s", queueName, uuid.New().String())
	}
	referenceUri = entry.ReferenceUri

	exist, err := rdb.Exists(ctx, referenceUri).Result()
	if err != nil && err != redis.Nil {
		return referenceUri, err
	}
	if exist == 1 {
		return referenceUri, nil
	}

	release := lock(rdb, referenceUri)
	defer func() {
		if err := release(); err != nil {
			log.Print(err)
		}
	}()

	err = rdb.Set(ctx, referenceUri, entry.String(), time.Hour*24).Err()
	if err != nil {
		return referenceUri, err
	}

	err = rdb.RPush(ctx, queueName, referenceUri).Err()
	if err != nil {
		return referenceUri, err
	}

	return referenceUri, nil
}

func internalPopFromTop(rdb *RedisClient, queueName string) (*XQueueEntry, error) {
	ctx := context.Background()

	cmd := rdb.BLPop(ctx, time.Second, queueName)
	v, err := cmd.Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return nil, nil
		}
		return nil, err
	}
	if len(v) < 2 {
		return nil, fmt.Errorf("invalid entry: %v", v)
	}
	referenceUri := v[1]
	release := lock(rdb, referenceUri)
	defer func() {
		if err := release(); err != nil {
			log.Print(err)
		}
	}()

	if !strings.HasPrefix(referenceUri, "gid://") {
		return nil, fmt.Errorf("invalid queue entry - should start with 'gid://': %v", v)
	}

	serializedEntry, err := rdb.Get(ctx, referenceUri).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get queue entry '%v': %v", referenceUri, err)
	}

	entry, err := parseQueueEntry(serializedEntry)
	if err != nil {
		return nil, fmt.Errorf("failed to parse queue entry: %v -> '%v'", err, serializedEntry)
	}

	if err = rdb.Del(ctx, referenceUri).Err(); err != nil {
		return nil, fmt.Errorf("failed to clean up queue entry: %v -> '%v'", err, serializedEntry)
	}

	return &entry, nil
}
