package tools

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/TwiN/go-color"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name       string
	bed        func(t *testing.T)
	sequential bool
	trace      string
}

type runnerFn func(cases []interface{})
type itFn func(name string, bed func(t *testing.T)) interface{}

func New(rootT *testing.T) (runnerFn, itFn) {
	cwd, _ := os.Getwd()
	cwdLen := len(cwd)

	getTestPath := func(testPath string) string {
		if cwdLen > 0 {
			if i := strings.Index(testPath, cwd); i == 0 {
				return testPath[cwdLen+1:]
			}
		}
		return testPath
	}
	formatCaller := func(pc uintptr, file string, line int, ok bool) string {
		if !ok {
			return "couldn't find the caller info"
		}
		return fmt.Sprintf("%s:%d", getTestPath(file), line)
	}

	var runner runnerFn = func(cases []interface{}) {
		rootT.Helper()

		for i := range cases {
			test, ok := cases[i].(testCase)
			if !ok {
				assert.Fail(rootT, fmt.Sprintf("invalid parameter passed to runner, requires return of 'it' but got %T", cases))
				continue
			}
			rootT.Run(test.name, func(t *testing.T) {
				name := test.name
				trace := test.trace
				if !test.sequential {
					t.Parallel()
				}
				defer func() {
					if t.Failed() {
						t.Log(color.Ize(color.Red, fmt.Sprintf(` --- FAIL: %s -> %s`, name, trace)))
					}
				}()
				defer func() {
					if r := recover(); r != nil {
						t.Logf(`!!! PANIC at %s -> %v`, formatCaller(runtime.Caller(4)), r)
						t.FailNow()
					}
				}()
				test.bed(t)
			})
		}
	}
	var it itFn = func(name string, bed func(t *testing.T)) interface{} {
		rootT.Helper()

		pc, file, no, ok := runtime.Caller(1)
		if !ok {
			assert.Fail(rootT, fmt.Sprintf("couldn't get a tracer for test case %s", name))
		}
		return testCase{
			name,
			bed,
			false,
			formatCaller(pc, file, no, ok),
		}

	}

	return runner, it
}
