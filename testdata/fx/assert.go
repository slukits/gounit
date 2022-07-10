// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fx

import (
	"fmt"
	"runtime"
	"strings"

	"github.com/slukits/gounit"
)

// TestTrueErrors implements a test calling t.True(false) and overwrites
// the default errorer.
type TestTrueErrors struct {
	FixtureLog
	gounit.Suite
	// Exp is logged from this suites Errorer implementation
	Exp string
}

func (s *TestTrueErrors) True_assertion_error(t *gounit.T) {
	t.True(false)
}

func (s *TestTrueErrors) error(args ...interface{}) {
	s.Logs = s.Exp
}

func (s *TestTrueErrors) Error() func(args ...interface{}) {
	return s.error
}

func (s *TestTrueErrors) File() string { return file }

// TestTrueReturnsFalse has a test logging the return-value of
// t.True(false).
type TestTrueReturnsFalse struct {
	FixtureLog
	gounit.Suite
}

func (s *TestTrueReturnsFalse) True_assertion_returns_false(t *gounit.T) {
	t.Log(t.True(false))
}

func (s *TestTrueReturnsFalse) error(args ...interface{}) {}

func (s *TestTrueReturnsFalse) Error() func(args ...interface{}) {
	return s.error
}

func (s *TestTrueReturnsFalse) File() string { return file }

// TestTrueReturnsFalse has a test logging the return-value of
// t.True(false).
type TestTrueReturnsTrue struct {
	FixtureLog
	gounit.Suite
}

func (s *TestTrueReturnsTrue) True_assertion_returns_true(t *gounit.T) {
	t.Log(t.True(true))
}

func (s *TestTrueReturnsTrue) error(args ...interface{}) {}

func (s *TestTrueReturnsTrue) Error() func(args ...interface{}) {
	return s.error
}

func (s *TestTrueReturnsTrue) File() string { return file }

type TestTrueError struct {
	FixtureLog
	gounit.Suite
	Msg string
	t   *gounit.T
}

func (s *TestTrueError) True_assertion_overwrites_error_msg(
	t *gounit.T,
) {
	s.t = t
	t.True(false, s.Msg)
}

func (s *TestTrueError) error(args ...interface{}) {
	s.t.Log(args...)
}

func (s *TestTrueError) Error() func(args ...interface{}) {
	return s.error
}

func (s *TestTrueError) File() string { return file }

type TestTrueFmtError struct {
	FixtureLog
	gounit.Suite
	Msgs []interface{}
	t    *gounit.T
}

func (s *TestTrueFmtError) True_assertion_overwrites_error_msg(
	t *gounit.T,
) {
	s.t = t
	t.True(false, s.Msgs...)
}

func (s *TestTrueFmtError) error(args ...interface{}) {
	s.t.Log(args...)
}

func (s *TestTrueFmtError) Error() func(args ...interface{}) {
	return s.error
}

func (s *TestTrueFmtError) File() string { return file }

type TestAssertion struct {
	FixtureLog
	gounit.Suite

	// True executes on given T-instance a successful assertion and
	// returns its value which is expected to be true
	True func(*gounit.T) bool

	// False executes on given T-instance a failing assertion and
	// returns its value which is expected to be false
	False func(*gounit.T) bool

	// Fails executes on given T-instance a failing assertion and
	// returns the expected error-message
	Fails func(*gounit.T) string

	// Overwrites executes on given T-instance a failing assertion with
	// an overwritten error message and returns the expected
	// error-message's suffix.
	Overwrite func(*gounit.T, string)

	// Overwrites executes on given T-instance a failing assertion with
	// an overwritten formatted error message and returns the expected
	// error-message's suffix.
	FmtOverwrite func(*gounit.T, string, string)

	Msg string

	Msgs map[string]string
	t    *gounit.T
}

func (s *TestAssertion) funcName(n string) string {
	idx := strings.LastIndex(n, ".")
	if idx < 0 {
		return n
	}
	return n[idx+1:]
}

func (s *TestAssertion) log(key, value string) {
	if s.Msgs == nil {
		s.Msgs = map[string]string{}
	}
	s.Msgs[key] = value
}

func (s *TestAssertion) error(args ...interface{}) {
	pc := make([]uintptr, 6)
	n := runtime.Callers(1, pc)
	if n < 5 {
		return
	}
	frames := runtime.CallersFrames(pc[4:])
	frame, ok := frames.Next()
	if ok && strings.HasPrefix( // testing for panicking adds a call
		s.funcName(frame.Function), "func") {
		frame, _ = frames.Next()
	}
	s.log(s.funcName(frame.Function), fmt.Sprint(args...))
}

func (s *TestAssertion) Error() func(args ...interface{}) {
	return s.error
}

func (s *TestAssertion) Test_true(t *gounit.T) {
	if !s.True(t) {
		s.Msg = "test assertion: true: returned false"
		t.FailNow()
	}
}

func (s *TestAssertion) Test_false(t *gounit.T) {
	if s.False(t) {
		s.Msg = "test assertion: false: returned true"
		t.FailNow()
	}
}

func (s *TestAssertion) Test_fail(t *gounit.T) {
	exp := s.Fails(t)
	if !strings.Contains(s.Msgs["Test_fail"], exp) {
		s.Msg = fmt.Sprintf(
			"fail: expect fail-msg to contain: '%s'; got '%s'",
			exp, s.Msgs["Test_fail"])
		t.FailNow()
	}
}

func (s *TestAssertion) Test_failing_overwrite(t *gounit.T) {
	const exp = "overwritten error"
	s.Overwrite(t, exp)
	if !strings.Contains(s.Msgs["Test_failing_overwrite"], exp) {
		s.Msg = fmt.Sprintf(
			"failing overwrite: expect fail-msg to contain: '%s'; got '%s'",
			exp, s.Msgs["Test_failing_overwrite"])
		t.FailNow()
	}
}

func (s *TestAssertion) Test_failing_fmt_overwrite(t *gounit.T) {
	const exp = "overwritten error"
	s.FmtOverwrite(t, "%s", exp)
	if !strings.Contains(s.Msgs["Test_failing_fmt_overwrite"], exp) {
		s.Msg = fmt.Sprintf(
			"failing fmt-overwrite: expect fail-msg to contain: '%s'; got '%s'",
			exp, s.Msgs["Test_failing_fmt_overwrite"])
		t.FailNow()
	}
}
