package testrun

import (
	"fmt"
	"runtime"
	"strings"
	"testing"

	"github.com/TwiN/go-color"
	"github.com/stretchr/testify/assert"
)

type testCase struct {
	cType caseType
	name  string
	bed   func(t *testing.T)
}

type TestBedFn func(t *testing.T)

type runner struct {
	rootT *testing.T
}

type caseType int

const (
	Regular caseType = iota
	Skip
	Must
)

func (_runner *runner) Run(testCases ...testCase) {
	_runner.rootT.Helper()

	mustIndexes := map[int]struct{}{}
	indexes := map[int]struct{}{}
	for i := range testCases {
		test := testCases[i]
		switch test.cType {
		case Must:
			mustIndexes[i] = struct{}{}
		case Regular:
			indexes[i] = struct{}{}
		case Skip:
		default:
		}
	}

	// If we have must running cases we skip the rest
	if len(mustIndexes) > 0 {
		indexes = mustIndexes
	}

	for i := range testCases {
		if _, ok := indexes[i]; !ok {
			continue
		}
		test := testCases[i]
		_runner.rootT.Run(test.name, test.bed)
	}
}

func (_runner *runner) ItSeq(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, false, Regular, bed)
}
func (_runner *runner) FItSeq(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, false, Must, bed)
}
func (_runner *runner) XItSeq(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, false, Skip, bed)
}

func (_runner *runner) It(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, true, Regular, bed)
}
func (_runner *runner) FIt(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, true, Must, bed)
}
func (_runner *runner) XIt(name string, bed TestBedFn) testCase {
	return _it(_runner.rootT, name, true, Skip, bed)
}

func _it(rootT *testing.T, name string, isParalell bool, cType caseType, bed TestBedFn) testCase {
	rootT.Helper()

	pc, file, no, ok := runtime.Caller(2)
	if !ok {
		assert.FailNow(rootT, fmt.Sprintf("couldn't get a tracer for test case %s", name))
	}
	trace := formatCaller(pc, file, no, ok)

	return testCase{
		cType,
		name,
		func(t *testing.T) {
			if isParalell {
				t.Parallel()
			}
			defer func() {
				if r := recover(); r != nil {
					t.Log(color.Ize(color.Cyan, color.Ize(color.Bold, fmt.Sprintf(` --- PANIC: %v`, r))))
					t.Log(color.Ize(color.Green, fmt.Sprintf(` --- %s`, identifyPanic())))
					t.Log(color.Ize(color.Yellow, fmt.Sprintf(` --- FAIL: %s -> %s`, name, trace)))
					t.FailNow()
					return
				}
				if t.Failed() {
					t.Log(color.Ize(color.Yellow, fmt.Sprintf(` --- FAIL: %s -> %s`, name, trace)))
					return
				}
			}()
			bed(t)
		},
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

func identifyPanic() string {
	stack := []string{}
	pc := make([]uintptr, 20)

	n := runtime.Callers(0, pc[:])
	if n == 0 {
		return "no stack available :("
	}

	pc = pc[:n]
	frames := runtime.CallersFrames(pc)

	for {
		f, more := frames.Next()
		name, file, line := f.Function, f.File, f.Line

		if !strings.HasPrefix(name, "runtime.") && !strings.HasPrefix(name, "testing.") && !strings.Contains(name, "testrun.") {
			stack = append(stack, fmt.Sprintf("%v:%v", file, line))
		}

		if !more {
			break
		}
	}

	return strings.Join(stack, "\n -> ")
}
