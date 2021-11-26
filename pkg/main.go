package main

import (
	"context"

	"github.com/aminpaks/go-streams/pkg/env"
	"github.com/aminpaks/go-streams/pkg/global"
	"github.com/aminpaks/go-streams/pkg/svr"
	"github.com/aminpaks/go-streams/pkg/xredis"
)

func main() {
	// Initialization of context for dependencies
	global.DependencyContext = context.Background()

	// Instantiate Redis client
	rdb, err := xredis.NewClient(env.Get("REDIS_URL", "redis://localhost:6379"))
	if err != nil {
		panic(err)
	}
	defer rdb.Close()
	global.DependencyContext = xredis.SetClientContext(global.DependencyContext, rdb)

	// Serving API
	svr.New(env.Get("PORT", "3100"))
}
