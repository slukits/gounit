// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"fmt"
	"testing"
	"time"
)

// T instances are passed to suite tests providing means for logging,
// assertion, failing, cancellation and concurrency-control for a test:
//
//	type MySuite { gounit.Suite }
//
//	func (s *MySuite) A_test(t *gounit.T) { t.Log("A_test run") }
//
//	func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t)}
type T struct {
	t        *testing.T
	tearDown func(*T)
	logger   func(...interface{})
	errorer  func(...interface{})
	canceler func()
	fs       *FS
}

// NewT wraps given go testing.T instance.
func NewT(t *testing.T) *T {
	return &T{
		t:        t,
		logger:   t.Log,
		errorer:  t.Error,
		canceler: t.FailNow,
	}
}

// GoT returns a pointer to wrapped testing.T instance which was created
// by the testing.T-runner of the suite-runner's testing.T instance.
func (t *T) GoT() *testing.T { return t.t }

// Log writes given arguments to set logger which defaults to the logger
// of wrapped *testing.T* instance.  The default may be overwritten by a
// suite-embedder implementing the SuiteLogging interface.
func (t *T) Log(args ...interface{}) {
	t.t.Helper()
	t.logger(args...)
}

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

// Error logs given arguments and flags test as failed but continues its
// execution.  t's errorer defaults to a Error-call of a wrapped
// testing.T instance and may be overwritten for a test-suite by
// implementing *SuiteErrorer*.
func (t *T) Error(args ...interface{}) {
	t.t.Helper()
	t.errorer(args...)
}

// Errorf logs given format-string leveraging fmt.Sprintf and flags test
// as failed but continues its execution.  t's errorer defaults to a
// Error-call of a wrapped testing.T instance and may be overwritten for
// a test-suite by implementing *SuiteErrorer*.
func (t *T) Errorf(format string, args ...interface{}) {
	t.t.Helper()
	t.Error(fmt.Sprintf(format, args...))
}

// FailNow cancels the execution of the test after a potential tear-down
// was called.  t's canceler defaults to a FailNow-call of a wrapped
// testing.T instance and may be overwritten for a test-suite by
// implementing *SuiteCanceler*.
func (t *T) FailNow() {
	t.t.Helper()
	if t.tearDown != nil {
		t.tearDown(t)
	}
	t.canceler()
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

// FatalOn cancels receiving test (see [T.FailNow]) after logging given
// error message iff passed argument is not nil and is a no-op
// otherwise.
func (t *T) FatalOn(err error) {
	t.t.Helper()
	if err == nil {
		return
	}
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

// Timeout returns a channel which is closed after given duration has
// elapsed.  Is given duration 0 it defaults to 10ms.
func (t *T) Timeout(d time.Duration) chan struct{} {
	if d == 0 {
		d = 10 * time.Millisecond
	}
	done := make(chan struct{})
	go func() {
		time.Sleep(d)
		close(done)
	}()
	return done
}

// FS returns an FS-instance with handy features for file system
// operations for testing.  I.e. copying a "golden" test file from a
// packages "testdata" directory to a test specific temporary directory
// looks like this:
//
//	t.FS().Data().Copy(golden, t.FS().Temp())
//
// It also removes error handling for file system operations by simply
// failing the test in case of an error.
func (t *T) FS() *FS {
	if t.fs == nil {
		t.fs = &FS{t: t}
	}
	return t.fs
}

// InitPrefix prefixes logging-messages of the Init-method to enable the
// reporter to discriminate Init-logs and Finalize-logs.
const InitPrefix = "__init__"

// FinalPrefix prefixes logging-messages of the Finalize-method to
// enable the reporter to discriminate Finalize-logs and Init-logs.
const FinalPrefix = "__final__"

// S instances are passed from gounit into a test-suite's Init-method or
// Finalize method, i.e. it is the T-instance of an Init/Finalize
// special method:
//
//	type MySuite { gounit.Suite }
//
//	func (s *MySuite) Init(t *gounit.S) { t.Log("init called") }
//
//	func (s *MySuite) MyTest(t *gounit.T) { t.Log("my test executed") }
//
//	func (s *MySuite) Finalize(t *gounit.S) { t.Log("finalize called") }
//
//	func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t) }
//
// An S instance provides logging-mechanisms and the possibility to
// cancel a suite's tests-run.  Note implementations of SuiteLogger or
// SuiteCanceler in a test-suite replace the default logging or
// cancellation behavior of an S-instance.  It defaults to testing.T.Log
// and testing.T.FailNow of the wrapped suit-runner's testing.T
// instance.
type S struct {
	t        *testing.T
	logger   func(...interface{})
	canceler func()
	prefix   string
}

// GoT returns a pointer to wrapped testing.T instance of the
// suite-runner's test.
func (st *S) GoT() *testing.T { return st.t }

// Log given arguments to wrapped test-runner's testing.T-logger or its
// replacement provided by a suite's SuiteLogger-implementation.
func (st *S) Log(args ...interface{}) {
	st.t.Helper()
	st.logger(append([]interface{}{st.prefix}, args...)...)
}

// Logf format logs leveraging fmt.Sprintf given arguments to wrapped
// test-runner's testing.T-logger or its replacement provided by a
// suite's SuiteLogger-implementation.
func (st *S) Logf(format string, args ...interface{}) {
	st.t.Helper()
	st.Log(fmt.Sprintf(format, args...))
}

// Fatal cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (st *S) Fatal(args ...interface{}) {
	st.t.Helper()
	st.Log(args...)
	st.canceler()
}

// Fatalf cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (st *S) Fatalf(format string, args ...interface{}) {
	st.t.Helper()
	st.Logf(format, args...)
	st.canceler()
}

// FatalOn cancels the test-suite's tests-run iff given error is not
// nil.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement provided by a
// suite's SuiteCanceler-implementation.
func (st *S) FatalOn(err error) {
	st.t.Helper()
	if err != nil {
		st.Fatal(err.Error())
	}
}
