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

	tt.FireRune('j')           // focus package
	tt.FireRune('j')           // focus go-test with subs
	tt.FireKey(tcell.KeyEnter) // and select it

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["go/pass: folded"]...,
	)
	t.Not.StarMatched(
		tt.Reporting().String(),
		fxNotExp["go/pass: folded"]...,
	)
}

func (s *Report) Go_tests_and_suites_are_initially_folded(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pass"]...,
	)

	for _, s := range fxNotExp["mixed/pass"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

func (s *Report) Unfolds_go_tests_on_folded_go_tests_selection(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pass"]...,
	)

	tt.FireRune('j')           // focus package
	tt.FireRune('j')           // focus go-tests
	tt.FireKey(tcell.KeyEnter) // and select it

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass go folded subs"]...,
	)
	for _, s := range fxNotExp["mixed/pass go folded subs"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

func (s *Report) Folds_go_tests_on_unfolded_go_tests_selection(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pass"]...,
	)

	tt.FireRune('j')           // focus package
	tt.FireRune('j')           // focus go-tests
	tt.FireKey(tcell.KeyEnter) // and select it

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass go folded subs"]...,
	)
	for _, s := range fxNotExp["mixed/pass go folded subs"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}

	tt.FireRune('j')           // focus package
	tt.FireRune('j')           // focus go-tests
	tt.FireKey(tcell.KeyEnter) // and select it

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass"]...,
	)
	for _, s := range fxNotExp["mixed/pass"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

func (s *Report) Unfolds_suite_tests_on_folded_suite_selection(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pass"]...,
	)

	// select second suite (pkg[0], blank[1], go[2], suite1[3], suite2[4])
	tt.ClickReporting(4)

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass second suite"]...,
	)
	for _, s := range fxNotExp["mixed/pass second suite"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

func (s *Report) Folds_suite_tests_on_unfolded_suite_selection(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pass"]...,
	)

	tt.ClickReporting(3) // select first suite (pkg, blank, go, suite)

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass first suite"]...,
	)
	for _, s := range fxNotExp["mixed/pass first suite"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}

	tt.ClickReporting(2) // select suite (pkg, blank, suite)

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass"]...,
	)
	for _, s := range fxNotExp["mixed/pass"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

func (s *Report) Unfolds_go_sub_tests_on_go_test_selection_(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pass"]...,
	)

	tt.ClickReporting(2) // select go tests
	tt.ClickReporting(7) // select first go test with sub-tests

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass go unfold"]...,
	)
	for _, s := range fxNotExp["mixed/pass go unfold"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

type dbg struct {
	Suite
	Fixtures
}

func (s *dbg) fxSource(t *T, dir string) (*lines.Events, *Testing) {
	return fxSourceDBG(t, s, dir)
}

func (s *dbg) Dbg(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pass"]...,
	)

	tt.ClickReporting(2) // select go tests
	tt.ClickReporting(7) // select first go test with sub-tests

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass go unfold"]...,
	)
	for _, s := range fxNotExp["mixed/pass go unfold"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

func TestDBG(t *testing.T) { Run(&dbg{}, t) }

func TestReport(t *testing.T) {
	t.Parallel()
	Run(&Report{}, t)
}
