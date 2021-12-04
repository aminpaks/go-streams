package re

import (
	"encoding/json"
	"net/http"
	"strings"

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

func JsonErrors(errs ...JsonObj) interface{} {
	return JsonObj{
		"errors": errs,
	}
}

func ToJsonError(err interface{}) JsonObj {
	switch v := err.(type) {
	case error:
		return JsonObj{"message": capitalizeFirstLetter(v.Error())}
	case string:
		return JsonObj{"message": v}
	case []byte:
		return JsonObj{"message": string(v)}
	case JsonObj:
		return v
	default:
	}
	return JsonObj{"message": "Unknown"}
}

func ToJsonErrors(errs ...interface{}) []JsonObj {
	list := []JsonObj{}
	for _, err := range errs {
		if err != nil {
			list = append(list, ToJsonError(err))
		}
	}
	return list
}

func capitalizeFirstLetter(i string) string {
	if len(i) > 1 {
		return strings.ToUpper(i[0:1]) + i[1:]
	}
	return i
}
