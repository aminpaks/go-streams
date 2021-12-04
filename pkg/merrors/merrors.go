package merrors

import (
	"encoding/json"

	"github.com/aminpaks/go-streams/pkg/re"
)

type Errorable interface {
	ErrorDetail() interface{}
}

func NewMerrors() *Merrors {
	return &Merrors{
		errors: make([]interface{}, 0),
	}
}

type Merrors struct {
	errors []interface{}
}

func (me *Merrors) Error() string {
	err := ""
	for i := range me.errors {
		err += getError(me.errors[i])
	}
	return err
}

func (me *Merrors) Has() bool {
	return me.errors != nil && len(me.errors) > 0
}

func (me *Merrors) Add(err error) *Merrors {
	if err != nil {
		me.errors = append(me.errors, err.Error())
	}
	return me
}

func (me *Merrors) AddCustom(err re.JsonObj) *Merrors {
	me.errors = append(me.errors, err)
	return me
}

func getError(i interface{}) string {
	switch v := i.(type) {
	case string:
		return v
	case error:
		return v.Error()
	default:
		if b, err := json.Marshal(v); err != nil {
			return string(b)
		}
		return "unknown error"
	}
}

func ErrorsOrElse(err error) []re.JsonObj {
	var errs []re.JsonObj
	switch v := err.(type) {
	case *Merrors:
		errs = re.ToJsonErrors(v.errors...)
	case error:
		errs = re.ToJsonErrors(v)
	default:
		errs = re.ToJsonErrors("Unknown error")
	}
	return errs
}
