// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

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
	if testSuite.Logs != "" {
		t.Fatal("expected initially an empty log")
	}
	Run(testSuite, t)
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
	t.Not.True(s.suiteErrorerUsed)
	t.Errorf("err")
	t.True(s.suiteErrorerUsed)
}

func (s *T_Instance) Times_out_after_given_duration(t *T) {
	d := 2 * time.Millisecond
	start := time.Now()
	<-t.Timeout(d)
	t.True(d <= time.Since(start))
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

func TestTrueAssertion(t *testing.T) {
	Run(&TrueAssertion{}, t)
}

type AssertionTests struct{ Suite }

func (s *AssertionTests) For_negated_true_assertion(t *T) {
	t.True(t.Not.True(false))
	t.Mock().Errorer(func(i ...interface{}) {})
	t.Not.True(t.Not.True(true))
}

func (s *AssertionTests) For_equality(t *T) {
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Eq(42, 42) },
		False: func(t *T) bool { return t.Eq(42, 22) },
		Fails: func(t *T) string {
			t.Eq(&(struct{ a int }{42}), struct{ a int }{42})
			return "assert equal"
		},
	}
	if !t.GoT().Run("AssertEquality", func(_t *testing.T) {
		Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_equality_of_pointers(t *T) {
	suite := &fx.PointerEqualityFX{}
	Run(suite, t.GoT())
	if suite.Logs != "" {
		t.Error(suite.Logs)
	}
}

func (s *AssertionTests) For_equality_of_strings(t *T) {
	suite := &fx.StringEqualityFX{}
	Run(suite, t.GoT())
	if suite.Logs != "" {
		t.Error(suite.Logs)
	}
}

func (s *AssertionTests) For_equality_of_stringer_implementations(
	t *T,
) {
	suite := &fx.TestStringerEquality{}
	Run(suite, t.GoT())
	if suite.Logs != "" {
		t.Error(suite.Logs)
	}
}

func (s *AssertionTests) For_equality_of_struct_fmt_strings(t *T) {
	suite := &fx.TestStructEquality{}
	Run(suite, t.GoT())
	if suite.Logs != "" {
		t.Error(suite.Logs)
	}
}

func (s *AssertionTests) For_inequality(t *T) {
	t.True(t.Not.Eq(&T{}, &T{}))
	t.True(t.Not.Eq(42, 22))
	t.True(t.Not.Eq(42, "42"))
	t.True(t.Not.Eq(struct{ n int }{n: 42}, struct{ n int }{n: 22}))
	t.True(t.Not.Eq(fx.TestStringer{Str: "22"}, fx.TestStringer{Str: "42"}))

	t.Mock().Errorer(func(i ...interface{}) {})
	t.Not.True(t.Not.Eq(t, t))
	t.Not.True(t.Not.Eq(42, 42))
	t.Not.True(t.Not.Eq("42", "42"))
	t.Not.True(t.Not.Eq(struct{ n int }{n: 42}, struct{ n int }{n: 42}))
	t.Not.True(t.Not.Eq(
		fx.TestStringer{Str: "42"}, fx.TestStringer{Str: "42"}))
}

func (s *AssertionTests) For_containing(t *T) {
	expErr := fmt.Sprintf(ContainsErr, "a\n", "\nb")
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Contains("a", "a") },
		False: func(t *T) bool { return t.Contains("a", "b") },
		Fails: func(t *T) string {
			t.Contains("a", "b")
			return expErr
		},
	}
	if !t.GoT().Run("AssertPanics", func(_t *testing.T) {
		Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_not_containing(t *T) {
	t.True(t.Not.Contains("22", "4"))
	t.Mock().Errorer(func(i ...interface{}) {})
	t.Not.True(t.Not.Contains("22", "2"))
}

func (s *AssertionTests) For_matched(t *T) {
	expErr := fmt.Sprintf(MatchedErr, "b", "a")
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Matched("a", "a") },
		False: func(t *T) bool { return t.Matched("a", "b") },
		Fails: func(t *T) string {
			t.Matched("a", "b")
			return expErr
		},
	}
	if !t.GoT().Run("AssertMatched", func(_t *testing.T) {
		Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_not_matching(t *T) {
	t.True(t.Not.Matched("22", "4"))
	t.Mock().Errorer(func(i ...interface{}) {})
	t.Not.True(t.Not.Matched("22", "2"))
}

func (s *AssertionTests) For_space_matched(t *T) {
	expErr := fmt.Sprintf(MatchedErr, "b", "a")
	suite := &fx.TestAssertion{
		True: func(t *T) bool {
			return t.SpaceMatched("a b", "a", "b")
		},
		False: func(t *T) bool {
			return t.SpaceMatched("a", "b")
		},
		Fails: func(t *T) string {
			t.SpaceMatched("a", "b")
			return expErr
		},
	}
	if !t.GoT().Run("AssertSpaceMatched", func(_t *testing.T) {
		Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_not_space_matched(t *T) {
	t.True(t.Not.SpaceMatched("22", "4", "2"))
	t.Mock().Errorer(func(i ...interface{}) {})
	t.Not.True(t.Not.SpaceMatched("22", "2", "2"))
}

func (s *AssertionTests) For_star_matched(t *T) {
	expErr := fmt.Sprintf(MatchedErr, "(?s)b", "a")
	suite := &fx.TestAssertion{
		True: func(t *T) bool {
			return t.StarMatched("a b", "a\nb")
		},
		False: func(t *T) bool {
			return t.StarMatched("a", "b")
		},
		Fails: func(t *T) string {
			t.StarMatched("a", "b")
			return expErr
		},
	}
	if !t.GoT().Run("AssertStarMatched", func(_t *testing.T) {
		Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_negated_star_match(t *T) {
	t.True(t.Not.StarMatched("22", "4", "2"))
	t.Mock().Errorer(func(i ...interface{}) {})
	t.Not.True(t.Not.StarMatched("42", "4", "2"))
}

func (s *AssertionTests) For_error(t *T) {
	suite := &fx.TestAssertion{
		True:  func(t *T) bool { return t.Err(errors.New("")) },
		False: func(t *T) bool { return t.Err(nil) },
		Fails: func(t *T) string {
			t.Err(nil)
			return ErrErr
		},
	}
	if !t.GoT().Run("AssertError", func(_t *testing.T) {
		Run(suite, _t)
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
			return ErrIsErr
		},
	}
	if !t.GoT().Run("AssertErrorIs", func(_t *testing.T) {
		Run(suite, _t)
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
			return ErrMatchedErr
		},
	}
	if !t.GoT().Run("AssertErrorIs", func(_t *testing.T) {
		Run(suite, _t)
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
			return PanicsErr
		},
	}
	if !t.GoT().Run("AssertPanics", func(_t *testing.T) {
		Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func (s *AssertionTests) For_within(t *T) {
	newTs := func(d uint) *TimeStepper {
		return (&TimeStepper{}).SetDuration(
			time.Duration(d) * time.Millisecond)
	}
	suite := &fx.TestAssertion{
		True: func(t *T) bool {
			return t.Within(newTs(2), func() bool { return true })
		},
		False: func(t *T) bool {
			return t.Within(newTs(2), func() bool { return false })
		},
		Fails: func(t *T) string {
			t.Within(newTs(2), func() bool { return false })
			return WithinErr
		},
	}
	if !t.GoT().Run("AssertWithin", func(_t *testing.T) {
		Run(suite, _t)
	}) {
		t.GoT().Fatalf("assertion suite failed: %s", suite.Msg)
	}
}

func TestAssertionTests(t *testing.T) {
	Run(&AssertionTests{}, t)
}
