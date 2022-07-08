// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit_test

import (
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
