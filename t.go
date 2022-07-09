// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"fmt"
	"testing"
)

// T wraps a *testing.T*-instance into a gounit.T instance which adjusts
// testing.T's api.
type T struct {
	Idx      int
	t        *testing.T
	tearDown func(*T)
	logger   func(...interface{})
	errorer  func(...interface{})
	cancler  func()
}

// Log writes given arguments to set logger which defaults to the logger
// of wrapped *testing.T* instance.  The default may be overwritten by a
// suite-embedder implementing the SuiteLogging interface.
func (t *T) Log(args ...interface{}) { t.logger(args...) }

// Logf writes given format string leveraging Sprintf to set logger which
// defaults to the logger of wrapped *testing.T* instance.  The default
// may be overwritten by a suite-embedder implementing the SuiteLogging
// interface.
func (t *T) Logf(format string, args ...interface{}) {
	t.Log(fmt.Sprintf(format, args...))
}

// Parallel signals that this test may be run in parallel with other
// parallel flagged tests.
func (t *T) Parallel() { t.t.Parallel() }

// Error flag test as failed but continue execution.
func (t *T) Error(args ...interface{}) {
	t.t.Helper()
	t.errorer(args...)
}

func (t *T) FailNow() {
	t.t.Helper()
	if t.tearDown != nil {
		t.tearDown(t)
	}
	t.cancler()
}

func (t *T) FatalIfNot(assertion bool) {
	if assertion {
		return
	}
	t.t.Helper()
	t.FailNow()
}

func (t *T) FatalOn(err error) {
	if err == nil {
		return
	}
	t.t.Helper()
	t.Fatal(err.Error())
}

func (t *T) Fatal(args ...interface{}) {
	t.t.Helper()
	t.Log(args...)
	t.FailNow()
}

func (t *T) Fatalf(format string, args ...interface{}) {
	t.t.Helper()
	t.Log(fmt.Sprintf(format, args...))
	t.FailNow()
}

const TrueErr = "expected given value to be true"

// True errors the test and returns false iff given value is not true;
// otherwise true is returned.  Given (formatted) message replace the
// default error message, i.e. msg[0] must be a string if len(msg) == 1
// it must be a format-string iff len(msg) > 1.
func (t *T) True(value bool, msg ...interface{}) bool {
	t.t.Helper()
	if true != value {
		t.Error(assertErr("true", TrueErr, msg...))
		return false
	}
	return true
}

const Assert = "assert %s: %v"
const FormatMsgErr = "expected first message-argument to be string; got %T"

func assertErr(label, msg string, args ...interface{}) string {
	if len(args) == 0 {
		return fmt.Sprintf(Assert, label, msg)
	}
	if len(args) == 1 {
		return fmt.Sprintf(Assert, label, args[0])
	}
	ext, ok := args[0].(string)
	if !ok {
		return fmt.Sprintf("%s: %s", fmt.Sprintf(Assert, label, msg),
			fmt.Sprintf(FormatMsgErr, args[0]))
	}
	return fmt.Sprintf(Assert, label, fmt.Sprintf(ext, args[1:]...))
}
