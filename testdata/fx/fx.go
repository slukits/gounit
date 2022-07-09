// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fx provides gounit test-fixture suites.
//
// Each test-fixture suite embeds the FixtureLog ensuring that all
// loggings during a suite's test runs are appended to the
// *Logs*-property which then can be evaluate after the suite's test
// runs.
package fx

import (
	"errors"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/slukits/gounit"
)

// FixtureLog provides the general logging facility for test suites
// fixtures by implementing gounit.SuiteLogger.
type FixtureLog struct{ Logs string }

// log logs given arguments to the *Logs* property
func (fl *FixtureLog) log(args ...interface{}) {
	fl.Logs += fmt.Sprint(args...)
}

// Logger implements the Logger interface, i.e. the suite-tests runner
// will use the returned function to implement gounit.T.Log/Logf.
func (fl *FixtureLog) Logger() func(args ...interface{}) {
	return fl.log
}

var file = func() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("fix: suites: can't determine file")
	}
	return f
}()

// TestAllSuiteTestsAreRun is a suite fixture to verify that the
// suite-test runner executes public suite-methods as tests.
type TestAllSuiteTestsAreRun struct {
	gounit.Suite
	FixtureLog
	// Exp is logged iff *A_test*-method is called
	Exp string
}

// A_test as a public method should be run by the suite-tests runner,
// i.e. log the content of *Exp*.
func (s *TestAllSuiteTestsAreRun) A_test(t *gounit.T) { t.Log(s.Exp) }

// private can't be run.
func (s *TestAllSuiteTestsAreRun) private(t *gounit.T) { t.Log("failed") }

func (fl *TestAllSuiteTestsAreRun) File() string { return file }

// TestSuiteLogging tests if a implemented SuiteLogger of a test-suite
// is used for logging.
type TestSuiteLogging struct {
	FixtureLog
	gounit.Suite
	// Exp is logged iff *Log_test*-is called
	Exp string
	// ExpFmt is logged if *Log_fmt_test*-is called
	ExpFmt string
}

// Log_test logs *Exp*.
func (s *TestSuiteLogging) Log_test(t *gounit.T) { t.Log(s.Exp) }

// Log_fmt_test logs *ExpFmt*.
func (s TestSuiteLogging) Log_fmt_test(t *gounit.T) {
	t.Logf("%s", s.ExpFmt)
}

func (fl *TestSuiteLogging) File() string { return file }

// TestIndexing logs for each test-call its name and index to evaluate
// if tests are indexed by their order of appearance.  This suite's
// tests are all run in parallel to ensure they don't run ordered.
type TestIndexing struct {
	gounit.Suite
	Exp   map[string]int
	Got   map[string]int
	mutex *sync.Mutex
}

func NewTestIndexingSuite(exp map[string]int) *TestIndexing {
	s := &TestIndexing{Exp: exp, mutex: &sync.Mutex{}}
	return s
}

// log interprets its first argument as test-method name, its second as
// its index and inserts it into *Got*.
func (s *TestIndexing) log(args ...interface{}) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if len(args) != 2 {
		panic("test indexing: log: expect exactly two arguments")
	}
	name, ok := args[0].(string)
	if !ok {
		panic("test indexing: log: expected first arg to be string")
	}
	idx, ok := args[1].(int)
	if !ok {
		panic("test indexing: log: expected second arg to be int")
	}
	if s.Got == nil {
		s.Got = make(map[string]int, 7)
	}
	s.Got[name] = idx
}

// Logger implements the Logger interface, i.e. the suite-tests runner
// will use the returned function to implement gounit.T.Log/-Logf.
func (s *TestIndexing) Logger() func(args ...interface{}) {
	return s.log
}

func (s *TestIndexing) Test_0(t *gounit.T) {
	t.Parallel()
	t.Log("Test_0", t.Idx)
}

func (s *TestIndexing) Test_1(t *gounit.T) {
	t.Parallel()
	t.Log("Test_1", t.Idx)
}

func (s *TestIndexing) Test_2(t *gounit.T) {
	t.Parallel()
	t.Log("Test_2", t.Idx)
}

func (s *TestIndexing) Test_3(t *gounit.T) {
	t.Parallel()
	t.Log("Test_3", t.Idx)
}

func (s *TestIndexing) Test_4(t *gounit.T) {
	t.Parallel()
	t.Log("Test_4", t.Idx)
}

func (s *TestIndexing) Test_5(t *gounit.T) {
	t.Parallel()
	t.Log("Test_5", t.Idx)
}

func (s *TestIndexing) Test_6(t *gounit.T) {
	t.Parallel()
	t.Log("Test_6", t.Idx)
}

func (fl *TestIndexing) File() string { return file }

// TestSetup  has its *SetUp*-method called before each test iff it logs
// "-11-22" or "-22-11" or "-1-212" or "-1-221" or "-2-121" or "-2-112".
// NOTE this suite's tests run in parallel making an effort to randomly
// pause a setup or test execution to have different log-values for
// different test-runs.
type TestSetup struct {
	FixtureLog
	gounit.Suite
}

func (s *TestSetup) SetUp(t *gounit.T) {
	t.Parallel()
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(-1 * (t.Idx + 1))
}

func (s *TestSetup) Test_A(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(t.Idx + 1)
}

func (s *TestSetup) Test_B(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(t.Idx + 1)
}

func (s *TestSetup) File() string {
	return file
}

// TestTearDown  has its *SetUp*-method called before each test iff it logs
// "1-12-2" or "2-21-1" or "12-1-2" or "12-2-1" or "21-2-1" or "21-1-2".
// NOTE this suite's tests run in parallel making an effort to randomly
// pause a tear-down or test execution to have different log-values for
// different test-runs.
type TestTearDown struct {
	FixtureLog
	gounit.Suite
}

func (s *TestTearDown) SetUp(t *gounit.T) {
	t.Parallel()
}
func (s *TestTearDown) TearDown(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(-1 * (t.Idx + 1))
}

func (s *TestTearDown) Test_A(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(t.Idx + 1)
}

func (s *TestTearDown) Test_B(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(t.Idx + 1)
}

func (s *TestTearDown) File() string { return file }

type TestTearDownAfterCancel struct {
	FixtureLog
	gounit.Suite
}

func (s *TestTearDownAfterCancel) TearDown(t *gounit.T) {
	t.Log(t.Idx)
}

func (s *TestTearDownAfterCancel) Fail_now_test(t *gounit.T) {
	t.FailNow() // should log 0
}

func (s *TestTearDownAfterCancel) Fatal_if_not_test(t *gounit.T) {
	t.FatalIfNot(false) // should log 1
}

func (s *TestTearDownAfterCancel) Fatal_on_test(t *gounit.T) {
	t.FatalOn(errors.New("")) // should log 2
}

func (s *TestTearDownAfterCancel) Fatal_test(t *gounit.T) {
	t.Fatal("") // should log 3
}

func (s *TestTearDownAfterCancel) Fatalf_test(t *gounit.T) {
	t.Fatalf("%s", "") // should log 4
}

func (s *TestTearDownAfterCancel) Cancel() func() {
	return func() {}
}

func (s *TestTearDownAfterCancel) File() string { return file }
