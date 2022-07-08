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

	"github.com/slukits/gounit"
)

// FixtureLog provides the general logging facility for test suites
// fixtures.
type FixtureLog struct{ Logs string }

// log logs given arguments to the *Logs* property
func (fl *FixtureLog) log(args ...interface{}) {
	fl.Logs += fmt.Sprint(args...)
}

// Logger implements the Logger interface, i.e. the suite-tests runner
// will use the returned function to implement gounit.T.Log.
func (fl *FixtureLog) Logger() func(args ...interface{}) {
	return fl.log
}

// TestAllSuiteTestsAreRun is a suite fixture to verify that the
// suite-test runner executes public suite-methods as tests.
type TestAllSuiteTestsAreRun struct {
	FixtureLog
	gounit.Suite
	// Exp is logged iff *A_test*-method is called
	Exp string
}

// A_test as a public method should be run by the suite-tests runner,
// i.e. log the content of *Exp*.
func (s *TestAllSuiteTestsAreRun) A_test(t *gounit.T) { t.Log(s.Exp) }

// private can't be run.
func (s *TestAllSuiteTestsAreRun) private(t *gounit.T) { t.Log("failed") }

type TestSuiteLogging struct {
	FixtureLog
	gounit.Suite
	// Exp is logged iff *Log_test*-is called
	Exp string
	// ExpFmt is logged if *Log_fmt_test*-is called
	ExpFmt string
}

// Log_test logs *Exp*.
func (s TestSuiteLogging) Log_test(t *gounit.T) { t.Log(s.Exp) }

// Log_fmt_test logs *ExpFmt*.
func (s TestSuiteLogging) Log_fmt_test(t *gounit.T) {
	t.Logf("%s", s.ExpFmt)
}
