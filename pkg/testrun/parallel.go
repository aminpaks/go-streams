package testrun

import "testing"

func P(fn func(t *testing.T)) func(t *testing.T) func(t *testing.T) {
	return func(t *testing.T) func(t *testing.T) {
		t.Parallel()

		return fn
	}
}
