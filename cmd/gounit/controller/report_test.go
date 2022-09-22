// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"strings"
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
		fxExp["go/pass suite"]...,
	)
	t.Not.StarMatched(
		tt.Reporting().String(),
		fxNotExp["go/pass suite"]...,
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

func (s *Report) Unfolds_go_sub_tests_on_go_test_selection(t *T) {
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

func (s *Report) fxSourceTouched(
	t *T, dir, touch string,
) (*lines.Events, *Testing) {
	return fxSourceTouched(t, s, dir, touch)
}

func (s *Report) Folded_packages_on_reported_package_selection(t *T) {
	_, tt := s.fxSourceTouched(t, "mixed/pp", "mixed/pp/pkg0")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pp/pkg0"]...,
	)

	tt.ClickReporting(0) // select package
	t.StarMatched(tt.Reporting().String(), fxExp["mixed/pp"]...)
}

func (s *Report) Selected_folded_package(t *T) {
	_, tt := s.fxSourceTouched(t, "mixed/pp", "mixed/pp/pkg0")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["mixed/pp/pkg0"]...,
	)
	tt.ClickReporting(0) // select package
	t.StarMatched(tt.Reporting().String(), fxExp["mixed/pp"]...)
	tt.ClickReporting(3) // select package 3

	t.SpaceMatched(
		tt.Reporting().String(),
		fxExp["mixed/pp/pkg3"]...,
	)
}

func (s *Report) Logged_text(t *T) {
	_, tt := s.fxSource(t, "logging")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["logging"]...,
	)

	tt.ClickReporting(2) // click on "go-tests"
	t.StarMatched(
		tt.Reporting().String(), fxExp["logging go-test"]...)

	// select the go-test-suite
	for i, l := range tt.Reporting() {
		if !strings.Contains(l.String(), "test go suite log") {
			continue
		}
		tt.ClickReporting(i)
		break
	}
	t.StarMatched(
		tt.Reporting().String(), fxExp["logging go-sub-test"]...)

	tt.ClickReporting(2) // back to folded view
	tt.ClickReporting(3) // select suite
	t.StarMatched(
		tt.Reporting().String(), fxExp["logging suite"]...)
}

func (s *Report) lineIsFailing(l lines.TestLine) bool {
	for i, r := range l.String() {
		if r == ' ' {
			continue
		}
		return l.Styles().Of(i).HasBG(tcell.ColorRed)
	}
	return false
}

func (s *Report) Failing_go_tests_ony_package(t *T) {
	failingLines := []int{0, 2, 7}
	_, tt := s.fxSource(t, "fail/gonly")
	got := tt.afterWatch(awReporting)
	t.FatalIfNot(t.True(len(got) > 7))
	for _, l := range failingLines {
		t.True(s.lineIsFailing(got[l]))
	}

	tt.ClickReporting(7)
	failingLines = []int{0, 2, 4}
	got = tt.Reporting()
	t.FatalIfNot(t.True(len(got) > 4))
	for _, l := range failingLines {
		t.True(s.lineIsFailing(got[l]))
	}
}

func (s *Report) Failing_package_due_to_compile_error(t *T) {
	_, tt := s.fxSource(t, "fail/compile")
	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["fail compile"]...,
	)
}

func TestReport(t *testing.T) {
	t.Parallel()
	Run(&Report{}, t)
}
