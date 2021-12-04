package users_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	miniredis "github.com/alicebob/miniredis/v2"
	"github.com/go-chi/chi/v5"
	"github.com/go-redis/redis"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

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

			ctrl, server := buildUserControllerTestMocks(t)
			defer ctrl.Finish()
			defer server.Close()

			resp := reqtest.New(server.URL).Do(t)
			body := reqtest.ReadResponse(t, resp.Body)
			defer resp.Body.Close()

			assert.Equal(http.StatusNotImplemented, resp.StatusCode)
			assert.Contains(body, "Not implemented")
		}),

		r.It("Should return success and stream user entity for post request", func(t *testing.T) {
			assert := assert.New(t)

			ctrl, server := buildUserControllerTestMocks(t)
			defer ctrl.Finish()
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
		}),
	)
}

func buildUserControllerTestMocks(t *testing.T) (*gomock.Controller, *httptest.Server) {
	ctrl := gomock.NewController(t)
	ctx := context.Background()

	mr, err := miniredis.Run()
	if err != nil {
		t.FailNow()
		return nil, nil
	}
	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	ctx = xredis.SetClientContext(ctx, redisClient)

	r := chi.NewRouter()
	err = users.NewUserController(ctx, r)
	if err != nil {
		t.FailNow()
		return nil, nil
	}

	ts := httptest.NewServer(r)

	return ctrl, ts
}
