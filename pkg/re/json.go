package re

import (
	"encoding/json"
	"net/http"

	"github.com/aminpaks/go-streams/pkg/h"
)

type JsonObj map[string]interface{}

func internalRenderJson(rw http.ResponseWriter, status int, v interface{}) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	rw.Header().Set(`content-type`, `application/json`)
	rw.WriteHeader(status)
	_, err = rw.Write(b)
	if err != nil {
		return err
	}
	return nil
}

func Json(status int, v interface{}) h.Renderer {
	return h.Renderer(func(rw http.ResponseWriter) error {
		return internalRenderJson(rw, status, v)
	})
}

func BuildJsonErrors(errs ...error) interface{} {
	list := []interface{}{}
	for _, err := range errs {
		if err != nil {
			list = append(list, map[string]string{
				"message": err.Error(),
			})
		}
	}

	return map[string]interface{}{
		"errors": list,
	}
}
