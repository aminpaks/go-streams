package h

import (
	"net/http"
)

type HandleFn func(rw http.ResponseWriter, r *http.Request) Renderer
type Renderer func(rw http.ResponseWriter) error

func NewH(fn HandleFn) http.HandlerFunc {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		dw := doneWriter{rw, false}

		renderer := fn(dw, r)
		err := renderer(dw)
		if err != nil && !dw.done {
			dw.WriteHeader(http.StatusInternalServerError)
			dw.Write([]byte(`Something is wrong!`))
		}
	})
}

type doneWriter struct {
	http.ResponseWriter
	done bool
}

func (w doneWriter) WriteHeader(status int) {
	w.done = true
	w.ResponseWriter.WriteHeader(status)
}

func (w doneWriter) Write(b []byte) (int, error) {
	w.done = true
	return w.ResponseWriter.Write(b)
}
