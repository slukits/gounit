// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"fmt"
	"testing"
	"time"

	"github.com/slukits/gounit/pkg/tfs"
)

// A TMock instance is obtained by [T.Mock] and provides the
// possibilities to mock logging, error handing and canceling of a test
// which default to [testing.T.Log], [testing.T.Error] and
// [testing.T.FailNow].
type TMock struct{ t *T }

// Logger is the final call which does the actual logging for [T.Log],
// [T.Error] and [T.Fatal]*(...), i.e. this function will receive all
// the log calls of these functions.
func (m *TMock) Logger(l func(...interface{})) { m.t.logger = l }

// Errorer the last function call of an error reporting function which
// by default reports back to the go testing framework indicating the
// failing of the test ... if mocked the later is prevented.
func (m *TMock) Errorer(e func(...interface{})) { m.t.errorer = e }

// Canceler the last function call of an error canceling function like
// Fatal which by default reports back to the go testing framework
// to stop the test execution instantly ... if mocked the later is
// prevented.
func (m *TMock) Canceler(c func()) { m.t.canceler = c }

// Reset sets logger, errorer and canceler back to their respective
// defaults.
func (m *TMock) Reset() {
	m.t.logger = m.t.GoT().Log
	m.t.errorer = m.t.GoT().Error
	m.t.canceler = m.t.GoT().FailNow
}

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
	fs       *tfs.FS

	// Not provides negations of T-assertions like Contains or StarMatched.
	Not Not
}

// NewT wraps given go testing.T instance into a gounit.T instance.
func NewT(t *testing.T) *T {
	_t := &T{
		t:        t,
		logger:   t.Log,
		errorer:  t.Error,
		canceler: t.FailNow,
	}
	_t.Not = Not{t: _t}
	return _t
}

// Mock provides the options to mock test logging, error handling and
// canceling.
func (t *T) Mock() *TMock {
	return &TMock{t: t}
}

// GoT returns a pointer to wrapped testing.T instance which usually was
// created by the testing.T-runner of the suite-runner's testing.T
// instance.
func (t T) GoT() *testing.T { return t.t }

// Log writes given arguments to set logger which defaults to the logger
// of wrapped testing.T instance.  The default is superseded by a
// suite-embedder implementing the [SuiteLogger] interface or by
// leveraging [T.Mock].
func (t T) Log(args ...interface{}) {
	t.t.Helper()
	t.logger(args...)
}

// Logf writes given format string to set logger which defaults to the
// logger of wrapped testing.T instance.  The default is superseded by a
// suite-embedder implementing the [SuiteLogger] interface or by
// leveraging [T.Mock].
func (t T) Logf(format string, args ...interface{}) {
	t.Log(fmt.Sprintf(format, args...))
}

// Parallel signals that this test may be run in parallel with other
// parallel flagged tests.
func (t T) Parallel() { t.t.Parallel() }

// Error logs given arguments and flags test as failed but continues its
// execution.  t's errorer defaults to a Error-call of a wrapped
// testing.T instance and may be overwritten for a test-suite by
// implementing [SuiteErrorer] or leveraging [T.Mock].
func (t T) Error(args ...interface{}) {
	t.t.Helper()
	t.errorer(args...)
}

// Errorf logs given format-string leveraging fmt.Sprintf and flags test
// as failed but continues its execution.  t's errorer defaults to a
// Error-call of a wrapped testing.T instance and may be overwritten for
// a test-suite by implementing [SuiteErrorer] or leveraging [T.Mock].
func (t T) Errorf(format string, args ...interface{}) {
	t.t.Helper()
	t.Error(fmt.Sprintf(format, args...))
}

// FailNow cancels the execution of the test after a potential tear-down
// was called.  t's canceler defaults to a FailNow-call of a wrapped
// testing.T instance and may be overwritten for a test-suite by
// implementing [SuiteCanceler] or leveraging [T.Mock].
func (t *T) FailNow() {
	t.t.Helper()
	if t.tearDown != nil {
		t.tearDown(t)
	}
	t.canceler()
}

// FatalIfNot cancels the test execution (see [T.FailNow]) if passed
// argument is false and is a no-op otherwise.
func (t T) FatalIfNot(assertion bool) {
	if assertion {
		return
	}
	t.t.Helper()
	t.FailNow()
}

// FatalOn cancels the test execution (see [T.FailNow]) after logging
// given error message iff passed argument is not nil and is a no-op
// otherwise.
func (t T) FatalOn(err error) {
	t.t.Helper()
	if err == nil {
		return
	}
	t.Fatal(err.Error())
}

// Fatal logs given arguments and cancels the test execution (see
// [T.FailNow]).
func (t T) Fatal(args ...interface{}) {
	t.t.Helper()
	t.Log(args...)
	t.FailNow()
}

// Fatalf logs given format-string leveraging fmt.Sprintf and cancels
// the test execution (see [T.FailNow]).
func (t T) Fatalf(format string, args ...interface{}) {
	t.t.Helper()
	t.Log(fmt.Sprintf(format, args...))
	t.FailNow()
}

// Timeout returns a channel which is closed after given duration has
// elapsed.  Is given duration 0 it defaults to 10ms.
func (t T) Timeout(d time.Duration) chan struct{} {
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
// package's "testdata" directory to a test specific temporary directory
// looks like this:
//
//	t.FS().Data().FileCopy(golden, t.FS().Temp())
//
// It also removes error handling for file system operations by simply
// failing the test in case of an error.
func (t *T) FS() *tfs.FS {
	if t.fs == nil {
		t.fs = tfs.New(t)
	}
	return t.fs
}

// InitPrefix prefixes logging-messages of the Init-method to enable the
// reporter to discriminate Init-logs and Finalize-logs.
const InitPrefix = "__init__"

// FinalPrefix prefixes logging-messages of the Finalize-method to
// enable the reporter to discriminate Finalize-logs and Init-logs.
const FinalPrefix = "__final__"

// S instances are passed from gounit into a test-suite's Init or
// Finalize method, i.e. it is the "T"-instance of an Init/Finalize
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
// cancel a suite's test-run.  Note implementations of [SuiteLogging] or
// [SuiteCanceler] in a test-suite replace the default logging or
// cancellation behavior of an S-instance.  NOTE an S-instance wraps the
// suite runner's [testing.T] instance while a T-instance wraps a suite
// runner's sub-test [testing.T] instance created for a particular suite
// test.
type S struct {
	t        *testing.T
	logger   func(...interface{})
	canceler func()
	prefix   string
	fs       *tfs.FS
}

// GoT returns a pointer to wrapped testing.T instance of the
// suite-runner's test.
func (st S) GoT() *testing.T { return st.t }

// Log given arguments to wrapped test-runner's testing.T-logger which
// is superseded by an optional [SuiteLogging]-implementation.
func (st S) Log(args ...interface{}) {
	st.t.Helper()
	st.logger(append([]interface{}{st.prefix}, args...)...)
}

// Logf format logs leveraging fmt.Sprintf given arguments to wrapped
// test-runner's testing.T-logger which is superseded by an optional
// [SuiteLogging]-implementation.
func (st S) Logf(format string, args ...interface{}) {
	st.t.Helper()
	st.Log(fmt.Sprintf(format, args...))
}

// Fatal cancels the test-suite's test-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance which is superseded by an optional
// [SuiteCanceler]-implementation.
func (st S) Fatal(args ...interface{}) {
	st.t.Helper()
	st.Log(args...)
	st.canceler()
}

// Fatalf cancels the test-suite's test-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance which is superseded by an optional
// [SuiteCanceler]-implementation.
func (st S) Fatalf(format string, args ...interface{}) {
	st.t.Helper()
	st.Logf(format, args...)
	st.canceler()
}

// FatalOn cancels the test-suite's test-run iff given error is not
// nil.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance which is superseded by an optional
// [SuiteCanceler]-implementation.
func (st S) FatalOn(err error) {
	st.t.Helper()
	if err != nil {
		st.Fatal(err.Error())
	}
}

// FS returns an FS-instance with handy features for file system
// operations for testing.  I.e. copying a "golden" test file from a
// packages "testdata" directory to a test specific temporary directory
// looks like this:
//
//	t.FS().Data().CopyFl(golden, t.FS().Temp())
//
// It also removes error handling for file system operations by simply
// failing the test in case of an error.
func (st *S) FS() *tfs.FS {
	if st.fs == nil {
		st.fs = tfs.New(st)
	}
	return st.fs
}
