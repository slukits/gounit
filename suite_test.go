// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit_test

import (
	"runtime"
	"strings"
	"testing"

	"github.com/slukits/gounit"
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

func Test_a_suite_s_tests_are_run(t *testing.T) {
	t.Parallel()
	testSuite := &fx.TestAllSuiteTestsAreRun{Exp: "A_test has been run"}
	if "" != testSuite.Logs {
		t.Fatal("expected initially an empty log")
	}
	gounit.Run(testSuite, t)
	if testSuite.Exp != testSuite.Logs {
		t.Errorf("expected test-suite log: %s; got: %s",
			testSuite.Exp, testSuite.Logs)
	}
}

func Test_a_suite_s_tests_are_indexed_by_appearance(t *testing.T) {
	t.Parallel()
	testSuite := fx.NewTestIndexingSuite(map[string]int{
		"Test_0": 0,
		"Test_1": 1,
		"Test_2": 2,
		"Test_3": 3,
		"Test_4": 4,
		"Test_5": 5,
		"Test_6": 6,
	})
	if testSuite.Got != nil {
		t.Fatal("expected initially empty *Got*-property")
	}
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result.
	if !t.Run("TestIndexing", func(_t *testing.T) {
		gounit.Run(testSuite, _t)
	}) {
		t.Fatalf("expected TestIndexing-suite to not fail")
	}
	if len(testSuite.Exp) != len(testSuite.Got) {
		t.Fatalf("expected %d logged tests; got: %d",
			len(testSuite.Exp), len(testSuite.Got))
	}
	for tst, idx := range testSuite.Exp {
		if testSuite.Got[tst] == idx {
			continue
		}
		t.Errorf("expected test %s to have index %d; got %d",
			tst, idx, testSuite.Got[tst])
	}
}

func Test_a_suite_s_file_is_its_test_file(t *testing.T) {
	t.Parallel()
	_, exp, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("couldn't determine test-file")
	}
	suite := &struct{ gounit.Suite }{}
	gounit.Run(suite, t)
	if exp != suite.File() {
		t.Errorf("expected suite file %s; got %s", exp, suite.File())
	}
}

type run struct{ gounit.Suite }

func (s *run) SetUp(t *gounit.T) { t.Parallel() }

func (s *run) Executes_setup_before_each_suite_test(t *gounit.T) {
	suite, goT := &fx.TestSetup{}, gounit.GoT(t)
	t.True(suite.Logs == "")
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result.
	if !goT.Run("TestSetup", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		goT.Fatalf("expected TestSetup-suite to not fail")
	}
	t.True(
		suite.Logs == "-11-22" || suite.Logs == "-22-11" ||
			suite.Logs == "-1-212" || suite.Logs == "-1-221" ||
			suite.Logs == "-2-121" || suite.Logs == "-2-112")
}

func (s *run) Executes_tear_down_after_each_suite_test(t *gounit.T) {
	suite, goT := &fx.TestTearDown{}, gounit.GoT(t)
	t.True(suite.Logs == "")
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result
	if !goT.Run("TestTearDown", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		goT.Fatalf("expected TestTearDown-suite to not fail")
	}
	t.True(
		suite.Logs == "1-12-2" || suite.Logs == "2-21-1" ||
			suite.Logs == "12-1-2" || suite.Logs == "12-2-1" ||
			suite.Logs == "21-2-1" || suite.Logs == "21-1-2")
}

func (s *run) Executes_tear_down_after_a_canceled_test(t *gounit.T) {
	suite, goT := &fx.TestTearDownAfterCancel{}, gounit.GoT(t)
	t.True(suite.Logs == "")
	gounit.Run(suite, goT)
	t.True("0011223344" == suite.Logs) // see suite's documentation
}

func (s *run) Executes_init_before_any_other_test(t *gounit.T) {
	suite, goT := &fx.TestInit{}, gounit.GoT(t)
	t.True(suite.Logs == "")
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result
	if !goT.Run("TestTearDown", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		goT.Fatalf("expected TestTearDown-suite to not fail")
	}
	t.True(10+len(gounit.InitPrefix) == len(suite.Logs))
	t.True(strings.HasPrefix(suite.Logs, gounit.InitPrefix))
}

func (s *run) Executes_finalize_after_all_test_ran(t *gounit.T) {
	suite, goT := &fx.TestFinalize{}, gounit.GoT(t)
	t.True(suite.Logs == "")
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result
	if !goT.Run("TestTearDown", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		goT.Fatalf("expected TestTearDown-suite to not fail")
	}
	t.True(10+len(gounit.FinalPrefix) == len(suite.Logs))
	t.True(strings.HasSuffix(suite.Logs, gounit.FinalPrefix))
}

func TestRun(t *testing.T) {
	t.Parallel()
	gounit.Run(&run{}, t)
}

type suite struct{ gounit.Suite }

func (s *suite) Canceler_implementation_overwrites_cancellation(
	t *gounit.T,
) {
	suite, goT := &fx.TestCancelerImplementation{}, gounit.GoT(t)
	t.True(suite.Got == nil)
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result
	if !goT.Run("TestTearDown", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		goT.Fatalf("expected TestTearDown-suite to not fail")
	}
	t.True(10 == len(suite.Got))
}

func TestSuite(t *testing.T) {
	t.Parallel()
	gounit.Run(&suite{}, t)
}

// type DBG struct{ gounit.Suite }
// func TestDBG(t *testing.T) { gounit.Run(&DBG{}, t) }
