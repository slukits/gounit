// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type Report struct {
	Suite
	Fixtures
}

func (s *Report) Init(t *S) {
	initGolden(t)
}

func (s *Report) SetUp(t *T) { t.Parallel() }

func (s *Report) TearDown(t *T) {
	fx := s.Get(t)
	if fx == nil {
		return
	}
	fx.(func())()
}

func (s *Report) fxSource(t *T, dir string) (*lines.Events, *Testing) {
	return fxSource(t, s, dir)
}

func (s *Report) Passing_go_tests_only(t *T) {
	_, tt := s.fxSource(t, "go/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["go/pass"]...,
	)
	t.StarMatched( // number of pkgs, suites, passed, failed
		tt.StatusBar().String(), "1", "2", "11", "0")
}

func (s *Report) Folds_tests_if_selecting_suite_with_shown_tests(t *T) {
	_, tt := s.fxSource(t, "go/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["go/pass"]...,
	)

	tt.FireRune('j')           // focus second
	tt.FireRune('j')           // focusable line (TestPass_4)
	tt.FireKey(tcell.KeyEnter) // and select it

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["go/pass: folded"]...,
	)

	str := tt.Reporting().String()
	t.Log(str)

}

func TestReport(t *testing.T) {
	t.Parallel()
	Run(&Report{}, t)
}
