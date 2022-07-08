// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit_test

import (
	"runtime"
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
