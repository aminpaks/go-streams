package xredis

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-redis/redis/v8"
)

const RedisClientContext = ContextKeyType("_RedisClient")

var ErrRedisConnect = errors.New("failed to connect")

func NewClient(redisUrl string) (*redis.Client, error) {
	clientOpt, err := redis.ParseURL(redisUrl)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRedisConnect, err)
	}

	ctx := context.Background()
	client := redis.NewClient(clientOpt)
	_, err = client.Ping(ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRedisConnect, err)
	}

	return client, nil
}

func GetClient(ctx context.Context) (*RedisClient, error) {
	if client, ok := ctx.Value(RedisClientContext).(*redis.Client); ok {
		return client, nil
	}
	return nil, errors.New("failed to get Redis Client from context")
}

func SetClientContext(ctx context.Context, client *redis.Client) context.Context {
	return context.WithValue(ctx, RedisClientContext, client)
}
