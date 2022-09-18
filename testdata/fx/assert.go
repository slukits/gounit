// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fx

import (
	"fmt"
	"regexp"
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

type PointerEqualityFX struct{ FX }

type T struct{ n int }

func (s *PointerEqualityFX) Test_equal_pointers(t *gounit.T) {
	if s.Logs != "" {
		return
	}
	testTypeInstance := T{n: 22}
	p1, p2 := &testTypeInstance, &testTypeInstance
	t.Eq(p1, p2)
}

func (s *PointerEqualityFX) Test_unequal_pointers(t *gounit.T) {
	if s.Logs != "" {
		return
	}
	t.Eq(&T{}, &T{})
	if !strings.HasPrefix(s.Logs, "assert equal: pointer:") {
		s.Logs = "pointer-equality: inequality: didn't get expected error"
		return
	}
	s.Logs = ""
}

var inequalityMatcher = regexp.MustCompile(
	`(?s)^assert equal: string-representations.*?[-].*?22.*?[+].*?42.*$`)

type StringEqualityFX struct{ FX }

func (s *StringEqualityFX) Test_equal_strings(t *gounit.T) {
	if s.Logs != "" {
		return
	}
	t.Eq("a", "a")
}

func (s *StringEqualityFX) Test_unequal_strings(t *gounit.T) {
	if s.Logs != "" {
		return
	}
	t.Eq("22", "42")
	if ok := inequalityMatcher.MatchString(s.Logs); !ok {
		s.Logs = "assert equal: strings: inequality: didn't get expected error"
		return
	}
	s.Logs = ""
}

type TestStructEquality struct{ FX }

func (s *TestStructEquality) Test_struct_equality(t *gounit.T) {
	if s.Logs != "" {
		return
	}
	t.Eq(struct{ n int }{n: 42}, struct{ n int }{n: 42})
}

func (s *TestStructEquality) Test_struct_inequality(t *gounit.T) {
	if s.Logs != "" {
		return
	}
	t.Eq(struct{ n int }{n: 22}, struct{ n int }{n: 42})
	if ok := inequalityMatcher.MatchString(s.Logs); !ok {
		s.Logs = "assert equal: struct: inequality: " +
			fmt.Sprintf("didn't get expected error: %s", s.Logs)
		return
	}
	s.Logs = ""
}

type TestStringerEquality struct {
	FX
}

type TestStringer struct{ Str string }

func (ts TestStringer) String() string { return ts.Str }

func (s *TestStringerEquality) Test_stringer_equality(t *gounit.T) {
	if s.Logs != "" {
		return
	}
	t.Eq(TestStringer{Str: "42"}, TestStringer{Str: "42"})
}

func (s *TestStringerEquality) Test_stringer_inequality(t *gounit.T) {
	if s.Logs != "" {
		return
	}
	t.Eq(TestStringer{Str: "22"}, TestStringer{Str: "42"})
	if ok := inequalityMatcher.MatchString(s.Logs); !ok {
		s.Logs = "assert equal: stringer: inequality: " +
			fmt.Sprintf("didn't get expected error: %s", s.Logs)
		return
	}
	s.Logs = ""
}

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

	msg string

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
	s.log(s.msg, fmt.Sprint(args...))
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
	s.msg = "Test_fail"
	exp := s.Fails(t)
	if !strings.Contains(s.Msgs["Test_fail"], exp) {
		s.Msg = fmt.Sprintf(
			"fail: expect fail-msg to contain: '%s'; got '%s'",
			exp, s.Msgs["Test_fail"])
		t.FailNow()
	}
}
