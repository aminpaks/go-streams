package users

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aminpaks/go-streams/pkg/h"
	"github.com/aminpaks/go-streams/pkg/merrors"
	"github.com/aminpaks/go-streams/pkg/re"
	"github.com/aminpaks/go-streams/pkg/xredis"
)

func NewUserController(depsCtx context.Context, r chi.Router) error {
	controller := &UserController{depsCtx: depsCtx}
	r.Post("/", h.New(controller.HandleCreate))
	r.Get("/", h.New(controller.HandleList))
	r.Get("/{userId}", h.New(controller.HandleGet))

	err := xredis.RegisterConsumer(depsCtx, "usersTest", "registerUsers", userCreationConsumer(), xredis.NewStreamConsumerOptions(2, 3))
	if err != nil {
		return fmt.Errorf("RegisterConsumer: %v", err)
	}

	return nil
}

type UserController struct {
	depsCtx    context.Context
	OnCreation func(user User)
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
