package engine

import "github.com/aminpaks/go-streams/pkg/xredis"

type EngineController struct {
	rdb *xredis.RedisClient
}

func NewEngineController(redisClient *xredis.RedisClient) *EngineController {
	return &EngineController{redisClient}
}
