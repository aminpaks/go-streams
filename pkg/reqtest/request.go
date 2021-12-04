package reqtest

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type HeaderKeyValue []string

type requestmaker struct {
	url     string
	headers []HeaderKeyValue
	method  string
	body    string
}

func (rm *requestmaker) Do(test *testing.T) *http.Response {
	return do(test, makeRequest(test, rm.method, rm.url, strings.NewReader(rm.body)))
}

func (rm *requestmaker) Body(body string) *requestmaker {
	rm.body = body
	return rm
}

func (rm *requestmaker) Method(method string) *requestmaker {
	rm.method = method
	return rm
}

func (rm *requestmaker) Headers(values ...HeaderKeyValue) *requestmaker {
	rm.headers = values
	return rm
}

func New(url string) *requestmaker {
	return &requestmaker{url: url}
}

func makeRequest(t *testing.T, method string, url string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		assert.FailNow(t, "failed to create http request: %v", err)
		return nil
	}
	return req
}

func ReadResponse(t *testing.T, r io.Reader) string {
	t.Helper()
	body, err := ioutil.ReadAll(r)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to read: %v", err))
		return ""
	}
	return string(body)
}
