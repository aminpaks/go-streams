package reqtest

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func do(t *testing.T, req *http.Request) *http.Response {
	t.Helper()
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		assert.FailNow(t, fmt.Sprintf("failed to make request: %v", err))
	}

	return resp
}
