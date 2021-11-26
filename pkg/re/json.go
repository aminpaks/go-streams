package re

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aminpaks/go-streams/pkg/h"
)

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

func BuildJsonMessage(msg string, v ...interface{}) interface{} {
	return map[string]interface{}{
		"message": fmt.Sprintf(msg, v...),
	}
}
