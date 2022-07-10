// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/slukits/gounit"
	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/testdata/fx"
)

// NOTE the here run tests create test-suite fixtures which are then run
// by the Run method using the tests testing.T instance.  This has the
// consequence that go test -v not only reports the tests of the
// test-files from this package but also the tests of test-suite
// fixtures.  The only way I could think of to avoid this would be to
// run the test-suite fixtures in its own "go test -v" system-call whose
// logged output then is evaluated.  But doing so would obscure the
// test-coverage which is also undesirable.

func Test_T_instance_logs_to_suite_s_logger(t *testing.T) {
	testSuite := &fx.TestSuiteLogging{Exp: "Log", ExpFmt: "Fmt"}
	if "" != testSuite.Logs {
		t.Fatal("expected initially an empty log")
	}
	gounit.Run(testSuite, t)
	if testSuite.Logs != "LogFmt" && testSuite.Logs != "FmtLog" {
		t.Errorf("expected test-suite log: LogFmt or FmtLog; got: %s",
			testSuite.Logs)
	}
}

type T_Instance struct {
	Suite
	suiteErrorerUsed bool
}

func SetUp(t *T) { t.Parallel() }

func (s *T_Instance) Noops_if_fatal_on_nil(t *T) {
	t.FatalOn(nil)
}

func (s *T_Instance) Noops_if_fatal_if_not_true(t *T) {
	t.FatalIfNot(true)
}

func (s *T_Instance) Error() func(...interface{}) {
	return func(i ...interface{}) { s.suiteErrorerUsed = true }
}

func (s *T_Instance) Uses_suite_s_errorer(t *T) {
	t.False(s.suiteErrorerUsed)
	t.Errorf("err")
	t.True(s.suiteErrorerUsed)
}

func TestTInstance(t *testing.T) {
	t.Parallel()
	Run(&T_Instance{}, t)
}

type TrueAssertion struct{ Suite }

func (s *TrueAssertion) Returns_true_if_passed_value_is_true(t *T) {
	suite := &fx.TestTrueReturnsTrue{}
	Run(suite, t.GoT())
	t.True(fmt.Sprintf("%v", true) == suite.Logs)
}

func (s *TrueAssertion) Returns_false_if_passed_value_is_false(t *T) {
	suite := &fx.TestTrueReturnsFalse{}
	Run(suite, t.GoT())
	t.True(fmt.Sprintf("%v", false) == suite.Logs)
}

func (s *TrueAssertion) Errors_if_passed_value_is_false(t *T) {
	suite := &fx.TestTrueErrors{Exp: "errorer called"}
	Run(suite, t.GoT())
	t.True(suite.Exp == suite.Logs)
}

func (s *TrueAssertion) Overwrites_error_message(t *T) {
	suite := &fx.TestTrueError{Msg: "replacement message"}
	Run(suite, t.GoT())
	t.True(strings.HasSuffix(suite.Logs, suite.Msg))
}

func (s *TrueAssertion) Overwrites_formatted_error_message(t *T) {
	msgs := []interface{}{"fmt %s", "error"}
	suite := &fx.TestTrueFmtError{Msgs: msgs}
	Run(suite, t.GoT())
	t.True(strings.HasSuffix(
		suite.Logs, fmt.Sprintf(msgs[0].(string), msgs[1])))
}

func (s *TrueAssertion) Fails_if_overwrites_have_no_string(t *T) {
	msgs := []interface{}{42, 42}
	suite := &fx.TestTrueFmtError{Msgs: msgs}
	Run(suite, t.GoT())
	t.True(strings.HasSuffix(
		suite.Logs, fmt.Sprintf(FormatMsgErr, msgs[0])))
}

func TestTrueAssertion(t *testing.T) {
	Run(&TrueAssertion{}, t)
}

type AssertionTests struct{ Suite }

func (s *AssertionTests) For_falsehood(t *T) {
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.False(false) },
		False: func(t *T) bool { return t.False(true) },
		Fails: func(t *T) string {
			t.False(true)
			return gounit.FalseErr
		},
		Overwrite: func(t *T, ow string) { t.False(true, ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.False(true, fmt, s)
		},
	}
	if !t.GoT().Run("AssertFalsehood", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_equality(t *T) {
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Eq(42, 42) },
		False: func(t *T) bool { return t.Eq(42, 22) },
		Fails: func(t *T) string {
			t.Eq(&(struct{ a int }{42}), struct{ a int }{42})
			return "assert equal"
		},
		Overwrite: func(t *T, ow string) { t.Eq(1, 2, ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.Eq("a", "b", fmt, s)
		},
	}
	if !t.GoT().Run("AssertEquality", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_inequality(t *T) {
	suite := &fx.TestAssertion{
		True: func(t *T) bool {
			return t.Neq(42, 22) && t.Neq("a", 42)
		},
		False: func(t *T) bool { return t.Neq(42, 42) },
		Fails: func(t *T) string {
			t.Neq(struct{ A int }{42}, struct{ A int }{42})
			return gounit.NeqErr
		},
		Overwrite: func(t *T, ow string) { t.Neq(1, 1, ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.Neq("a", "a", fmt, s)
		},
	}
	if !t.GoT().Run("AssertInequality", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_containing(t *T) {
	expErr := fmt.Sprintf(gounit.ContainsErr, "a\n", "\nb")
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Contains("a", "a") },
		False: func(t *T) bool { return t.Contains("a", "b") },
		Fails: func(t *T) string {
			t.Contains("a", "b")
			return expErr
		},
		Overwrite: func(t *T, ow string) { t.Contains("a", "b", ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.Contains("a", "b", fmt, s)
		},
	}
	if !t.GoT().Run("AssertPanics", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_matched(t *T) {
	expErr := fmt.Sprintf(gounit.MatchedErr, "b", "a")
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Matched("a", "a") },
		False: func(t *T) bool { return t.Matched("a", "b") },
		Fails: func(t *T) string {
			t.Matched("a", "b")
			return expErr
		},
		Overwrite: func(t *T, ow string) { t.Matched("a", "b", ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.Matched("a", "b", fmt, s)
		},
	}
	if !t.GoT().Run("AssertMatched", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_space_matched(t *T) {
	expErr := fmt.Sprintf(gounit.MatchedErr, "b", "a")
	suite := &fx.TestAssertion{
		True: func(t *T) bool {
			return t.SpaceMatched("a b", []string{"a", "b"})
		},
		False: func(t *T) bool {
			return t.SpaceMatched("a", []string{"b"})
		},
		Fails: func(t *T) string {
			t.SpaceMatched("a", []string{"b"})
			return expErr
		},
		Overwrite: func(t *T, ow string) {
			t.SpaceMatched("a", []string{"b"}, ow)
		},
		FmtOverwrite: func(t *T, fmt, s string) {
			t.SpaceMatched("a", []string{"b"}, fmt, s)
		},
	}
	if !t.GoT().Run("AssertSpaceMatched", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_star_matched(t *T) {
	expErr := fmt.Sprintf(gounit.MatchedErr, "(?s)b", "a")
	suite := &fx.TestAssertion{
		True: func(t *T) bool {
			return t.StarMatched("a b", []string{"a\nb"})
		},
		False: func(t *T) bool {
			return t.StarMatched("a", []string{"b"})
		},
		Fails: func(t *T) string {
			t.StarMatched("a", []string{"b"})
			return expErr
		},
		Overwrite: func(t *T, ow string) {
			t.StarMatched("a", []string{"b"}, ow)
		},
		FmtOverwrite: func(t *T, fmt, s string) {
			t.StarMatched("a", []string{"b"}, fmt, s)
		},
	}
	if !t.GoT().Run("AssertStarMatched", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_error(t *T) {
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Err(errors.New("")) },
		False: func(t *T) bool { return t.Err(nil) },
		Fails: func(t *T) string {
			t.Err(nil)
			return gounit.ErrErr
		},
		Overwrite: func(t *T, ow string) { t.Err(nil, ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.Err(nil, fmt, s)
		},
	}
	if !t.GoT().Run("AssertError", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_error_is(t *T) {
	err := errors.New("target")
	wrap := fmt.Errorf("wrap %w", err)
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.ErrIs(wrap, err) },
		False: func(t *T) bool { return t.ErrIs(err, errors.New("")) },
		Fails: func(t *T) string {
			t.ErrIs(nil, nil)
			return gounit.ErrIsErr
		},
		Overwrite: func(t *T, ow string) { t.ErrIs(nil, nil, ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.ErrIs(nil, nil, fmt, s)
		},
	}
	if !t.GoT().Run("AssertErrorIs", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_error_matched(t *T) {
	err := errors.New("err")
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.ErrMatched(err, "err") },
		False: func(t *T) bool { return t.ErrMatched(err, "42") },
		Fails: func(t *T) string {
			t.ErrMatched(nil, "")
			return gounit.ErrMatchedErr
		},
		Overwrite: func(t *T, ow string) { t.ErrMatched(nil, "", ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.ErrMatched(nil, "", fmt, s)
		},
	}
	if !t.GoT().Run("AssertErrorIs", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_panicking(t *T) {
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Panics(func() { panic("") }) },
		False: func(t *T) bool { return t.Panics(func() {}) },
		Fails: func(t *T) string {
			t.Panics(func() {})
			return gounit.PanicsErr
		},
		Overwrite: func(t *T, ow string) { t.Panics(func() {}, ow) },
		FmtOverwrite: func(t *T, fmt, s string) {
			t.Panics(func() {}, fmt, s)
		},
	}
	if !t.GoT().Run("AssertPanics", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func TestAssertionTests(t *testing.T) {
	Run(&AssertionTests{}, t)
}

type DBG struct{ Suite }

func TestDBG(t *testing.T) { Run(&DBG{}, t) }
