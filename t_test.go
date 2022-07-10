// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit_test

import (
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

type T_Instance struct{ Suite }

func SetUp(t *T) { t.Parallel() }

func (s *T_Instance) Noops_if_fatal_on_nil(t *T) {
	t.FatalOn(nil)
}

func (s *T_Instance) Noops_if_fatal_if_not_true(t *T) {
	t.FatalIfNot(true)
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

// type DBG struct{ Suite }
//
// func TestDBG(t *testing.T) { Run(&DBG{}, t) }
