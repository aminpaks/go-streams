package users

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"github.com/aminpaks/go-streams/pkg/global"
	"github.com/aminpaks/go-streams/pkg/h"
	"github.com/aminpaks/go-streams/pkg/re"
	"github.com/aminpaks/go-streams/pkg/xredis"
)

func NewUserController(r chi.Router) error {
	controller := &UserController{}
	r.Post("/", h.NewH(controller.HandleCreate))
	r.Get("/", h.NewH(controller.HandleList))
	r.Get("/{userName}", h.NewH(controller.HandleGet))

	err := xredis.RegisterConsumer(global.DependencyContext, "usersTest", "registerUsers", UserConsumer(), xredis.NewStreamConsumerOptions(2, 3))
	if err != nil {
		return fmt.Errorf("RegisterConsumer: %v", err)
	}

	return nil
}

type UserController struct {
	OnCreation func(user User)
}

func (us *UserController) HandleList(rw http.ResponseWriter, r *http.Request) h.Renderer {
	return re.Json(http.StatusNotImplemented, re.BuildJsonMessage("Not implemented yet!"))
}

func (us *UserController) HandleGet(rw http.ResponseWriter, r *http.Request) h.Renderer {
	return re.Json(http.StatusNotImplemented, re.BuildJsonMessage("Not implemented yet!"))
}

func (us *UserController) HandleCreate(rw http.ResponseWriter, r *http.Request) h.Renderer {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return re.Json(http.StatusInternalServerError, re.BuildJsonErrors(fmt.Errorf("failed to read request body")))
	}

	user, err := ParseUserWithOptionalId(body)
	if err != nil {
		return re.Json(http.StatusBadRequest, re.BuildJsonErrors(fmt.Errorf("invalid request payload: %v", err)))
	}

	ref, err := xredis.StreamAppend(global.DependencyContext, "usersTest", user.WithId(uuid.New()).String())
	if err != nil {
		log.Fatalf("failed to append entry to stream: %v", err)
		return re.Json(http.StatusInternalServerError, re.BuildJsonErrors(fmt.Errorf("failed to process request")))
	}

	return re.Json(http.StatusOK, re.BuildJsonMessage("User will be created shortly, reference: %v", ref))
}
