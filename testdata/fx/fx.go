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
	"fmt"
	"runtime"

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
func (s *TestSuiteLogging) Log_fmt_test(t *gounit.T) {
	t.Logf("%s", s.ExpFmt)
}

func (fl *TestSuiteLogging) File() string { return file }

// TestIndexing logs for each test-call its name and index to evaluate
// if tests are indexed by their order of appearance.
type TestIndexing struct {
	gounit.Suite
	Exp map[string]int
	Got map[string]int
}

// log interprets its first argument as test-method name, its second as
// its index and inserts it into *Got*.
func (s *TestIndexing) log(args ...interface{}) {
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
		s.Got = make(map[string]int, 3)
	}
	s.Got[name] = idx
}

// Logger implements the Logger interface, i.e. the suite-tests runner
// will use the returned function to implement gounit.T.Log/-Logf.
func (s *TestIndexing) Logger() func(args ...interface{}) {
	return s.log
}

func (s *TestIndexing) Test_0(t *gounit.T) {
	t.Log("Test_0", t.Idx)
}

func (s *TestIndexing) Test_1(t *gounit.T) {
	t.Log("Test_1", t.Idx)
}

func (s TestIndexing) Test_2(t *gounit.T) {
	t.Log("Test_2", t.Idx)
}

func (fl *TestIndexing) File() string { return file }
