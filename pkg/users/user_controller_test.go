package users_test

import (
	"context"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"

	"github.com/aminpaks/go-streams/pkg/async"
	"github.com/aminpaks/go-streams/pkg/reqtest"
	"github.com/aminpaks/go-streams/pkg/testrun"
	"github.com/aminpaks/go-streams/pkg/users"
	"github.com/aminpaks/go-streams/pkg/xredis"
)

func TestUserController(t *testing.T) {
	t.Parallel()

	r := testrun.New(t)

	r.Run(
		r.It("Should return not implemented in the response for get request", func(t *testing.T) {
			assert := assert.New(t)

			server, sg, shutdown := buildUserControllerTestMocks(t)
			defer server.Close()

			resp := reqtest.New(server.URL).Do(t)
			body := reqtest.ReadResponse(t, resp.Body)
			defer resp.Body.Close()

			assert.Equal(http.StatusNotImplemented, resp.StatusCode)
			assert.Contains(body, "Not implemented")

			shutdown()
			sg.WaitTimeout(time.Second * 10)
		}),

		r.It("Should return success and stream user entity for post request", func(t *testing.T) {
			assert := assert.New(t)

			server, sg, shutdown := buildUserControllerTestMocks(t)
			defer server.Close()

			resp := reqtest.New(server.URL).
				Method(http.MethodPost).
				Body(`{"name": "John Doe","email":"johndoe@me.com"}`).
				Headers([]string{"content-type", "application/json"}).
				Do(t)
			body := reqtest.ReadResponse(t, resp.Body)
			defer resp.Body.Close()

			assert.Equal(http.StatusOK, resp.StatusCode)
			assert.Contains(body, "User will be created shortly")

			shutdown()
			sg.WaitTimeout(time.Second)
		}),

		r.FIt("Should return success and stream user entity for post request", func(t *testing.T) {
			assert := assert.New(t)

			server, sg, shutdown := buildUserControllerTestMocks(t)
			defer server.Close()

			resp := reqtest.New(server.URL).
				Method(http.MethodPost).
				Body(`{"name": "John Doe","email":"johndoe@me.com"}`).
				Headers([]string{"content-type", "application/json"}).
				Do(t)
			body := reqtest.ReadResponse(t, resp.Body)
			defer resp.Body.Close()

			assert.Equal(http.StatusOK, resp.StatusCode)
			assert.Contains(body, "User will be created shortly")

			start := time.Now()
			shutdown()
			sg.WaitTimeout(time.Second)
			log.Printf("Dur: %s", time.Since(start))
		}),
	)
}

func buildUserControllerTestMocks(t *testing.T) (*httptest.Server, *async.SyncGroup, func()) {
	mr, err := miniredis.Run()
	if err != nil {
		t.FailNow()
		return nil, nil, nil
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	ctx := context.Background()
	ctx = xredis.SetClientContext(ctx, redisClient)

	shutdown, cancel := context.WithCancel(context.Background())
	r := chi.NewRouter()
	sg := async.NewSyncGroup()
	err = users.NewUserController(ctx, shutdown, sg, r)
	if err != nil {
		cancel()
		t.FailNow()
		return nil, nil, nil
	}

	ts := httptest.NewServer(r)

	return ts, sg, cancel
}
