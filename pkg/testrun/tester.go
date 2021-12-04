package testrun

import (
	"fmt"
	"runtime"
	"testing"

	"github.com/TwiN/go-color"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	name     string
	bed      func(t *testing.T)
	parallel bool
	skip     bool
	trace    string
}

type TestBedFn func(t *testing.T)

type runner struct {
	rootT *testing.T
}

func (_runner *runner) Run(testCases ...testCase) {
	_runner.rootT.Helper()

	for i := range testCases {
		test := testCases[i]
		_runner.rootT.Run(test.name, func(t *testing.T) {
			if test.skip {
				t.SkipNow()
			}
			if test.parallel {
				t.Parallel()
			}
			name := test.name
			trace := test.trace
			defer func() {
				if t.Failed() {
					t.Log(color.Ize(color.Yellow, fmt.Sprintf(` --- FAIL: %s -> %s`, name, trace)))
				}
			}()
			test.bed(t)
		})
	}
}
func (_runner *runner) ItSeq(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, false, false, bed)
}
func (_runner *runner) It(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, true, false, bed)
}
func (_runner *runner) XIt(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, true, true, bed)
}

func _it(rootT *testing.T, name string, isParalell bool, shouldSkip bool, bed TestBedFn) testCase {
	rootT.Helper()

	pc, file, no, ok := runtime.Caller(2)
	if !ok {
		assert.FailNow(rootT, fmt.Sprintf("couldn't get a tracer for test case %s", name))
	}
	return testCase{
		name,
		bed,
		isParalell,
		shouldSkip,
		formatCaller(pc, file, no, ok),
	}
}

func New(rootT *testing.T) *runner {
	return &runner{rootT}
}

func formatCaller(pc uintptr, file string, line int, ok bool) string {
	if !ok {
		return ""
	}
	return fmt.Sprintf("%s:%d", file, line)
}
