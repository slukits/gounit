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
//     type MySuite { gounit.Suite }
//
//     func (s *MySuite) A_test(t *gounit.T) { t.Log("A_test run") }
//
//     func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t)}
type T struct {
	t        *testing.T
	tearDown func(*T)
	logger   func(...interface{})
	errorer  func(...interface{})
	canceler func()
}

// GoT returns a pointer to wrapped testing.T instance which was created
// by the testing.T-runner of the suite-runner's testing.T instance.
func (t *T) GoT() *testing.T { return t.t }

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

// FatalOn cancels receiving test (see *FailNow*) after logging given
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

// InitPrefix prefixes logging-messages of the Init-method to enable the
// reporter to discriminate Init-logs and Finalize-logs.
const InitPrefix = "__init__"

// FinalPrefix prefixes logging-messages of the Finalize-method to
// enable the reporter to discriminate Finalize-logs and Init-logs.
const FinalPrefix = "__final__"

// I instances are passed from gounit into a test-suite's Init-method:
//
//     type MySuite { gounit.Suite }
//
//     func (s *MySuite) Init(t *gounit.I) { t.Log("init called") }
//
//     func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t) }
//
// An I instance provides logging-mechanisms and the possibility to
// cancel a suite's tests-run.  NOTE implementations of SuiteLogger or
// SuiteCanceler in a test-suite replace the default logging or
// cancellation behavior of an I-instance.  It defaults to testing.T.Log
// and testing.T.FailNow of the wrapped testing.T instance which is the
// one from the test-runner.
type I struct {
	t        *testing.T
	logger   func(...interface{})
	canceler func()
}

// GoT returns a pointer to wrapped testing.T instance of the
// suite-runner's test.
func (i *I) GoT() *testing.T { return i.t }

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
	i.canceler()
}

// Fatalf cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (i *I) Fatalf(format string, args ...interface{}) {
	i.t.Helper()
	i.Logf(format, args...)
	i.canceler()
}

// FatalOn cancels the test-suite's tests-run iff given error is not
// nil.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement provided by a
// suite's SuiteCanceler-implementation.
func (i *I) FatalOn(err error) {
	i.t.Helper()
	if err != nil {
		i.Fatal(err.Error())
	}
}

// F instances are passed from gounit into a test-suite's
// Finalize-method:
//
//     type MySuite { gounit.Suite }
//
//     func (s *MySuite) Finalize(t *gounit.F) {
//         t.Log("finalize called")
//     }
//
//     func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t) }
//
// An F instance provides logging-mechanisms and the possibility to
// cancel a suite's tests-run.  NOTE implementations of SuiteLogger or
// SuiteCanceler in a test-suite replace the default logging or
// cancellation behavior of an I-instance.  It defaults to testing.T.Log
// and testing.T.FailNow of the wrapped testing.T instance which is the
// one from the test-runner.
type F struct {
	t        *testing.T
	logger   func(...interface{})
	canceler func()
}

// GoT returns a pointer to wrapped testing.T instance of the
// suite-runner's test.
func (f *F) GoT() *testing.T { return f.t }

// Log given arguments to wrapped test-runner's testing.T-logger or its
// replacement provided by a suite's SuiteLogger-implementation.
func (f *F) Log(args ...interface{}) {
	f.t.Helper()
	f.logger(append([]interface{}{FinalPrefix}, args...)...)
}

// Logf format logs leveraging fmt.Sprintf given arguments to wrapped
// test-runner's testing.T-logger or its replacement provided by a
// suite's SuiteLogger-implementation.
func (f *F) Logf(format string, args ...interface{}) {
	f.t.Helper()
	f.Log(fmt.Sprintf(format, args...))
}

// Fatal cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (f *F) Fatal(args ...interface{}) {
	f.t.Helper()
	f.Log(args...)
	f.canceler()
}

// Fatalf cancels the test-suite's tests-run after given arguments were
// logged.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (f *F) Fatalf(format string, args ...interface{}) {
	f.t.Helper()
	f.Logf(format, args...)
	f.canceler()
}

// FatalOn cancels the test-suite's tests-run iff given error is not
// nil.  The cancellation defaults to a FailNow call of wrapped
// test-runner's testing.T-instance or its replacement  provided by a
// suite's SuiteCanceler-implementation.
func (f *F) FatalOn(err error) {
	f.t.Helper()
	if err != nil {
		f.Fatal(err.Error())
	}
}

// Timeout returns a channel which receives a message after given
// duration *d*
func (t *T) Timeout(d time.Duration) chan struct{} {
	done := make(chan struct{})
	go func() {
		time.Sleep(d)
		close(done)
	}()
	return done
}
