// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"fmt"
	"testing"
)

// T instances are passed to suite tests providing means for logging,
// assertion, failing, cancellation and concurrency-control for a test:
//
//      type MySuite { gounit.Suite }
//
//      func (s *MySuite) A_test(t *gounit.T) { t.Log("A_test run") }
//
//      func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t)}
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
// may be overwritten by a suite-embedder implementing the SuiteLogger
// interface.
func (t *T) Logf(format string, args ...interface{}) {
	t.Log(fmt.Sprintf(format, args...))
}

// Parallel signals that this test may be run in parallel with other
// parallel flagged tests.
func (t *T) Parallel() { t.t.Parallel() }

// Error log given arguments and flag test as failed but continue its
// execution.  t's errorer defaults to a Error-call of a wrapped
// testing.T instance and may be overwritten for a test-suite by
// implementing SuiteErrorer.
func (t *T) Error(args ...interface{}) {
	t.t.Helper()
	t.errorer(args...)
}

// FailNow cancels the execution of the test after a potential tear-down
// was called.  t's canceler defaults to a FailNow-call of a wrapped
// testing.T instance and may be overwritten for a test-suite by
// implementing SuiteCanceler.
func (t *T) FailNow() {
	t.t.Helper()
	if t.tearDown != nil {
		t.tearDown(t)
	}
	t.cancler()
}

// FatalIfNot cancels receiving test (see *FailNow*) if passed argument
// is false and is a no-op otherwise.
func (t *T) FatalIfNot(assertion bool) {
	if assertion {
		return
	}
	t.t.Helper()
	t.FailNow()
}

// FatalIfNot cancels receiving test (see *FailNow*) after logging given
// error message iff passed argument is not nil and is a no-op
// otherwise.
func (t *T) FatalOn(err error) {
	if err == nil {
		return
	}
	t.t.Helper()
	t.Fatal(err.Error())
}

// Fatal logs given arguments and cancels the test execution (see
// *FailNow*).
func (t *T) Fatal(args ...interface{}) {
	t.t.Helper()
	t.Log(args...)
	t.FailNow()
}

// Fatalf logs given format-string leveraging fmt.Sprintf and cancels
// the test execution (see *FailNow*).
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

const InitPrefix = "__init__"
const FinalPrefix = "__final__"

// I instances are passed from gounit into a test-suite's Init-method:
//
//      type MySuite { gounit.Suite }
//
//      func (s *MySuite) Init(t *gounit.I) { t.Log("init called") }
//
//      func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t) }
//
// An I instance provides logging-mechanisms and the possibility to
// cancel a suite's tests-run.  NOTE implementations of SuiteLogger or
// SuiteCanceler in a test-suite replace the default logging or
// cancellation behavior of an I-instance.  It defaults to testing.T.Log
// and testing.T.FailNow of the wrapped testing.T instance which is the
// one from the test-runner.
type I struct {
	t       *testing.T
	logger  func(...interface{})
	cancler func()
}

// Log given arguments to wrapped test-runner's testing.T-logger or its
// replacement provided by a suite's SuiteLogger-implementation.
func (i *I) Log(args ...interface{}) {
	i.t.Helper()
	i.logger(append([]interface{}{InitPrefix}, args...)...)
}

// Logf format logs leveraging fmt.Sprintf given arguments to wrapped
// test-runner's testing.T-logger or its replacement provided by a
// suite's SuiteLogger-implementation.
func (i *I) Logf(format string, args ...interface{}) {
	i.t.Helper()
	i.Log(fmt.Sprintf(format, args...))
}

// Fatal cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (i *I) Fatal(args ...interface{}) {
	i.t.Helper()
	i.Log(args...)
	i.cancler()
}

// Fatalf cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (i *I) Fatalf(format string, args ...interface{}) {
	i.t.Helper()
	i.Logf(format, args...)
	i.cancler()
}

// Fatalf cancels the test-suite's tests-run iff given error is not nil.
// The cancellation defaults to a FailNow call of wrapped test-runner's
// testing.T-instance or its replacement  provided by a suite's
// SuiteCanceler-implementation.
func (i *I) FatalOn(err error) {
	i.t.Helper()
	if err != nil {
		i.Fatal(err.Error())
	}
}

// F instances are passed from gounit into a test-suite's
// Finalize-method:
//
//      type MySuite { gounit.Suite }
//
//      func (s *MySuite) Finalize(t *gounit.I) {
//	 		t.Log("finalize called")
// 		}
//
//      func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t) }
//
// An I instance provides logging-mechanisms and the possibility to
// cancel a suite's tests-run.  NOTE implementations of SuiteLogger or
// SuiteCanceler in a test-suite replace the default logging or
// cancellation behavior of an I-instance.  It defaults to testing.T.Log
// and testing.T.FailNow of the wrapped testing.T instance which is the
// one from the test-runner.
type F struct {
	t       *testing.T
	logger  func(...interface{})
	cancler func()
}

// Log given arguments to wrapped test-runner's testing.T-logger or its
// replacement provided by a suite's SuiteLogger-implementation.
func (i *F) Log(args ...interface{}) {
	i.t.Helper()
	i.logger(append([]interface{}{FinalPrefix}, args...)...)
}

// Logf format logs leveraging fmt.Sprintf given arguments to wrapped
// test-runner's testing.T-logger or its replacement provided by a
// suite's SuiteLogger-implementation.
func (i *F) Logf(format string, args ...interface{}) {
	i.t.Helper()
	i.Log(fmt.Sprintf(format, args...))
}

// Fatal cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (i *F) Fatal(args ...interface{}) {
	i.t.Helper()
	i.Log(args...)
	i.cancler()
}

// Fatalf cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (i *F) Fatalf(format string, args ...interface{}) {
	i.t.Helper()
	i.Logf(format, args...)
	i.cancler()
}

// Fatalf cancels the test-suite's tests-run iff given error is not nil.
// The cancellation defaults to a FailNow call of wrapped test-runner's
// testing.T-instance or its replacement  provided by a suite's
// SuiteCanceler-implementation.
func (i *F) FatalOn(err error) {
	i.t.Helper()
	if err != nil {
		i.Fatal(err.Error())
	}
}
