package users

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aminpaks/go-streams/pkg/async"
	"github.com/aminpaks/go-streams/pkg/h"
	"github.com/aminpaks/go-streams/pkg/merrors"
	"github.com/aminpaks/go-streams/pkg/re"
	"github.com/aminpaks/go-streams/pkg/throttler"
	"github.com/aminpaks/go-streams/pkg/xredis"
)

func NewUserController(depsCtx context.Context, shutdown context.Context, syncGroup *async.SyncGroup, r chi.Router) error {
	queuer, err := xredis.BuildQueuer(depsCtx)
	if err != nil {
		return err
	}

	controller := &UserController{
		depsCtx:   depsCtx,
		enqueue:   queuer,
		queueName: "production/usersCreationQueue",
	}

	r.Post("/", h.New(controller.HandleCreate))
	r.Get("/", h.New(controller.HandleList))
	r.Get("/{userId}", h.New(controller.HandleGet))

	err = xredis.RegisterStreamConsumer(depsCtx, "usersTest", "registerUsers", userCreationConsumer(), xredis.NewStreamConsumerOptions(2, 3))
	if err != nil {
		return fmt.Errorf("RegisterConsumer: %v", err)
	}

	redisClient, err := xredis.GetClient(depsCtx)
	if err != nil {
		return err
	}

	worker := throttler.NewThrottler(shutdown)
	worker.Initialize()

	syncGroup.AddChannel(
		"test queue consumer",
		xredis.NewSortedQueueConsumerWithOptions(
			redisClient,
			shutdown,
			"test",
			testQueueEntryConsumer(worker),
			testQueueFailureHandler(),
			&xredis.XSortedQueueOptions{
				MaxRetries: 3,
				Consuming:  2,
				Consumers:  2,
			},
		),
	)

	// // Adding some samples to "test" priority queue
	// go func() {
	// 	rand.Seed(rand.Int63())
	// 	kind := "testingItems"
	// 	for i := 0; i < 3; i++ {
	// 		priority := rand.Float64()
	// 		if err := xredis.EnqueueSortedEntry(
	// 			redisClient,
	// 			context.Background(),
	// 			"test",
	// 			xredis.NewXSortedQueueEntry(fmt.Sprintf("check this out #%v", i+1), priority, xredis.NewUri(kind), time.Hour),
	// 		); err != nil {
	// 			log.Printf("ERROR Enqueue: %v", err)
	// 		}
	// 	}
	// }()

	return nil
}

type UserController struct {
	depsCtx    context.Context
	OnCreation func(user User)
	enqueue    xredis.XQueuer
	queueName  string
}

func (us *UserController) HandleList(rw http.ResponseWriter, r *http.Request) h.Renderer {
	return re.Json(
		http.StatusNotImplemented,
		re.JsonObj{
			"data": re.JsonObj{"message": "Not implemented yet!"},
		},
	)
}

func (us *UserController) HandleGet(rw http.ResponseWriter, r *http.Request) h.Renderer {
	return re.Json(
		http.StatusNotImplemented,
		re.JsonObj{
			"data": re.JsonObj{"message": "Not implemented!"},
		})
}

func (us *UserController) HandleCreate(rw http.ResponseWriter, r *http.Request) h.Renderer {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return re.Json(http.StatusInternalServerError, re.JsonErrors(re.ToJsonError("Failed to read request body")))
	}

	user, err := ParseUserWithOptionalId(body)
	if err != nil {
		return re.Json(http.StatusBadRequest, re.JsonErrors(re.ToJsonError(fmt.Sprintf("Invalid request payload: %v", err))))
	}
	if err := user.Validate(); err != nil {
		return re.Json(http.StatusBadRequest, re.JsonErrors(merrors.ErrorsOrElse(err)...))
	}

	ref, err := xredis.StreamAppend(us.depsCtx, "usersTest", user.WithId(uuid.New()).String())
	if err != nil {
		log.Printf("failed to append entry to stream: %v", err)
		return re.Json(http.StatusInternalServerError, re.JsonErrors(re.ToJsonError("Failed to process request")))
	}

	return re.Json(http.StatusOK, re.JsonObj{
		"data": re.JsonObj{
			"message":   "User will be created shortly",
			"reference": ref,
		},
	})
}
