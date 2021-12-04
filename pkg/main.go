package main

import (
	"context"

	"github.com/aminpaks/go-streams/pkg/env"
	"github.com/aminpaks/go-streams/pkg/svr"
	"github.com/aminpaks/go-streams/pkg/xredis"
)

func main() {
	// Initialization of context for dependencies
	dependencyContext := context.Background()

	// Instantiate Redis client
	rdb, err := xredis.NewClient(env.Get("REDIS_URL", "redis://localhost:6379"))
	if err != nil {
		panic(err)
	}
	defer rdb.Close()
	dependencyContext = xredis.SetClientContext(dependencyContext, rdb)

	// Serving API
	svr.New(dependencyContext, env.Get("PORT", "3100"))
}
