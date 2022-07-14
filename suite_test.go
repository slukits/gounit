// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit_test

import (
	"fmt"
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
	if testSuite.Logs != "" {
		t.Fatal("expected initially an empty log")
	}
	gounit.Run(testSuite, t)
	if testSuite.Exp != testSuite.Logs {
		t.Errorf("expected test-suite log: %s; got: %s",
			testSuite.Exp, testSuite.Logs)
	}
}

type run struct{ gounit.Suite }

func (s *run) SetUp(t *gounit.T) { t.Parallel() }

func (s *run) Executes_setup_before_each_suite_test(t *gounit.T) {
	suite := &fx.TestSetup{}
	t.True(suite.Logs == "")
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result.
	if !t.GoT().Run("TestSetup", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("expected TestSetup-suite to not fail")
	}
	t.True(
		suite.Logs == "-11-22" || suite.Logs == "-22-11" ||
			suite.Logs == "-1-212" || suite.Logs == "-1-221" ||
			suite.Logs == "-2-121" || suite.Logs == "-2-112")
}

func (s *run) Executes_tear_down_after_each_suite_test(t *gounit.T) {
	suite := &fx.TestTearDown{}
	t.True(suite.Logs == "")
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result
	if !t.GoT().Run("TestTearDown", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("expected TestTearDown-suite to not fail")
	}
	t.True(
		suite.Logs == "1-12-2" || suite.Logs == "2-21-1" ||
			suite.Logs == "12-1-2" || suite.Logs == "12-2-1" ||
			suite.Logs == "21-2-1" || suite.Logs == "21-1-2")
}

func (s *run) Executes_tear_down_after_a_canceled_test(t *gounit.T) {
	suite := &fx.TestTearDownAfterCancel{}
	t.True(suite.Logs == "")
	gounit.Run(suite, t.GoT())
	t.True(suite.Logs == "12345")
}

func (s *run) Executes_init_before_any_other_test(t *gounit.T) {
	suite := &fx.TestInit{}
	t.True(suite.Logs == "")
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result
	if !t.GoT().Run("TestTearDown", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("expected TestTearDown-suite to not fail")
	}
	t.True(10+len(gounit.InitPrefix) == len(suite.Logs))
	t.True(strings.HasPrefix(suite.Logs, gounit.InitPrefix))
}

func (s *run) Executes_finalize_after_all_test_ran(t *gounit.T) {
	suite := &fx.TestFinalize{}
	t.True(suite.Logs == "")
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result
	if !t.GoT().Run("TestTearDown", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("expected TestTearDown-suite to not fail")
	}
	t.True(10+len(gounit.FinalPrefix) == len(suite.Logs))
	t.True(strings.HasSuffix(suite.Logs, gounit.FinalPrefix))
}

func (s *run) Provides_its_test_to_init_and_finalize(t *gounit.T) {
	suite := &fx.TestInitFinalHaveRunTest{InitLog: "i", FinalLog: "f"}
	t.True(suite.Logs == "")
	if !t.GoT().Run("TestTearDown", func(_t *testing.T) {
		suite.RunT = _t
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("expected TestTearDown-suite to not fail: %s",
			suite.Fatal)
	}
	t.True(
		fmt.Sprintf("%si%sf", gounit.InitPrefix, gounit.FinalPrefix) ==
			suite.Logs)
}

func TestRun(t *testing.T) {
	t.Parallel()
	gounit.Run(&run{}, t)
}

type suite struct{ gounit.Suite }

func (s *suite) Canceler_implementation_overwrites_cancellation(
	t *gounit.T,
) {
	suite := &fx.TestCancelerImplementation{}
	t.True(suite.Got == nil)
	// run testSuite in a sub-test to ensure all its tests are run
	// before we investigate the result
	if !t.GoT().Run("TestTearDown", func(_t *testing.T) {
		gounit.Run(suite, _t)
	}) {
		t.GoT().Fatalf("expected TestTearDown-suite to not fail")
	}
	t.True(len(suite.Got) == 10)
}

func TestSuite(t *testing.T) {
	t.Parallel()
	gounit.Run(&suite{}, t)
}
