// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fx

import "github.com/slukits/gounit"

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
