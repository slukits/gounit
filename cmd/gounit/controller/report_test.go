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

func (s *Report) fxSourceTouched(
	t *T, dir, touch string,
) (*lines.Events, *Testing) {

	return fxSourceTouched(t, s, dir, touch)
}

func (s *Report) Go_tests_only(t *T) {
	_, tt := s.fxSource(t, "go/pass")

	t.StarMatched(tt.Reporting().String(), fxExp["go/pass"]...)
	t.StarMatched( // number of pkgs, suites, passed, failed
		tt.StatusBar().String(), "1", "2", "11", "0")
}

func (s *Report) Initially_most_recently_modified_package_folded(t *T) {
	_, tt := s.fxSourceTouched(t, "mixed/pass", "mixed/pass/suite3_test.go")

	t.StarMatched(
		tt.Reporting().String(),
		fxExp["mixed/pass init"]...,
	)

	for _, s := range fxNotExp["mixed/pass"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

func (s *Report) Logged_text(t *T) {
	_, tt := s.fxSource(t, "logging")

	tt.ClickReporting(3)
	t.StarMatched(tt.Reporting().String(), fxExp["logging suite"]...)

	tt.ClickReporting(2) // go to folded view
	t.StarMatched(
		tt.Reporting().String(), fxExp["logging folded"]...)

	tt.ClickReporting(2) // go to go-tests
	t.StarMatched(
		tt.Reporting().String(), fxExp["logging go-tests"]...)

	tt.ClickReporting(8) // go to go-suite
	t.StarMatched(
		tt.Reporting().String(), fxExp["logging go-suite"]...)
}

const expTxt = "Lorem ipsum dolor sit amet, consectetur adipiscing " +
	"elit. Morbi id mi rutrum, pretium ipsum et, gravida dui. " +
	"Vestibulum et sapien et diam interdum gravida sit amet quis " +
	"leo. Suspendisse ac nisi sit amet erat eleifend bibendum. Sed " +
	"eu tincidunt arcu, sit amet pretium arcu. Nam urna eros, " +
	"aliquet sed mi vitae, consectetur consequat purus. Donec " +
	"tincidunt dictum velit, at dictum quam tincidunt ut. " +
	"Pellentesque vel dolor lacinia, dictum justo sit amet, " +
	"bibendum ex. Maecenas sit amet pellentesque leo."

func (s *Report) Overlong_log_text_wrapped(t *T) {
	_, tt := s.fxSource(t, "wrapped")
	tt.ClickReporting(2)
	got := strings.ReplaceAll(tt.Reporting().String(), "\n", "")
	got = strings.ReplaceAll(got, " ", "")
	exp := strings.ReplaceAll(expTxt, " ", "")
	t.Contains(got, exp)
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

func (s *Report) Failing_go_tests_only_package(t *T) {
	failingLines := []int{0, 2, 8}
	_, tt := s.fxSource(t, "fail/gonly")
	got := tt.Reporting()
	t.FatalIfNot(t.True(len(got) > 8))
	for _, l := range failingLines {
		t.True(s.lineIsFailing(got[l]))
	}

	tt.ClickReporting(8)
	failingLines = []int{0, 2, 4}
	got = tt.Reporting()
	t.FatalIfNot(t.True(len(got) > 4))
	for _, l := range failingLines {
		t.True(s.lineIsFailing(got[l]))
	}
}

func (s *Report) Failing_package_due_to_compile_error(t *T) {
	_, tt := s.fxSource(t, "fail/compile")
	t.StarMatched(tt.Reporting().String(), fxExp["fail compile"]...)
}

func (s *Report) Failing_package_s_failing_go_tests_initially(t *T) {
	failingLines := []int{0, 2, 5, 10}
	_, tt := s.fxSource(t, "fail/mixed")

	got := tt.Reporting()
	t.FatalIfNot(t.True(len(got) > 10))
	for _, l := range failingLines {
		t.True(s.lineIsFailing(got[l]))
	}
}

func (s *Report) Always_failing_package_s_failing_suites(t *T) {
	failingLines := []int{13, 14}
	_, tt := s.fxSource(t, "fail/mixed")

	got := tt.Reporting()
	t.FatalIfNot(t.True(len(got) > 15))
	for _, l := range failingLines {
		t.True(s.lineIsFailing(got[l]))
	}
}

func (s *Report) Failing_go_suite_test(t *T) {
	_, tt := s.fxSource(t, "fail/mixed")

	tt.ClickReporting(10)
	t.StarMatched(tt.Reporting().String(), "go-tests", "p4 sub 2")

	exp, got := 0, 0
	for _, f := range fxExp["fail mixed go-suite"] {
		exp++
		for _, l := range tt.Reporting() {
			if strings.HasPrefix(strings.TrimSpace(l.String()), f) {
				got++
				s.lineIsFailing(l)
			}
		}
	}
	t.Eq(exp, got)
}

func (s *Report) Failing_suite_test_along_failing_suites_and_go_suite(
	t *T,
) {
	_, tt := s.fxSource(t, "fail/mixed")

	tt.ClickReporting(14)
	t.Contains(tt.Reporting().String(), "suite test 4 3")

	exp, got := 0, 0
	for _, f := range fxExp["fail mixed suite"] {
		exp++
		for _, l := range tt.Reporting() {
			if strings.HasPrefix(strings.TrimSpace(l.String()), f) {
				got++
				s.lineIsFailing(l)
			}
		}
	}
	t.Eq(exp, got)
}

func (s *Report) Failing_packages_always(t *T) {
	_, tt := s.fxSource(t, "fail/pp")
	t.Contains(tt.Reporting().String(), "fail/pp/fail1")
	t.Contains(tt.Reporting().String(), "fail/pp/fail2")
	tt.collapseAll()
	t.StarMatched(
		tt.Reporting().String(), fxExp["fail pp collapsed"]...)

	fx1, fx2 := "fail/pp/fail2", "fail/pp/pass"
	for i, l := range tt.Reporting() {
		if !strings.HasPrefix(l.String(), fx1) {
			continue
		}
		tt.ClickReporting(i)
	}
	t.Not.Contains(tt.Reporting().String(), fx2)
	t.StarMatched(tt.Reporting().String(), fxExp["fail pp"]...)

	tt.collapseAll()
	for i, l := range tt.Reporting() {
		if !strings.HasPrefix(l.String(), fx2) {
			continue
		}
		tt.ClickReporting(i)
	}
	t.StarMatched(tt.Reporting().String(), fxExp["fail pp"]...)
}

func (s *Report) Panic_during_test_execution_as_package_error(t *T) {
	_, tt := s.fxSource(t, "panic")
	t.StarMatched(
		tt.Reporting().String(),
		fxExp["panic"]...,
	)
}

func (s *Report) Current_package_vetted_if_vet_is_turned_on(t *T) {
	_, tt := s.fxSource(t, "vet")
	tt.ClickReporting(2) // unfold suite
	t.Contains(tt.Reporting().String(), "fails if vetted")
	t.Contains(tt.ButtonBar().String(), "[v]et=off")

	tt.beforeView(func() { tt.ClickButton("vet=off") })
	t.Contains(tt.ButtonBar().String(), "[v]et=on")
	t.Contains(tt.Trim(tt.Reporting()).String(), "FAIL")
	t.Contains(tt.Trim(tt.Reporting()).String(), "vet/src.go:11:26")
}

func (s *Report) Selected_package_vetted_if_vet_is_turned_on(t *T) {
	_, tt := s.fxSource(t, "vet")
	tt.ClickReporting(0)
	t.Not.Contains(tt.Trim(tt.Reporting()).String(), "fails if vetted")

	tt.ClickButton("vet=off")
	t.Contains(tt.ButtonBar().String(), "[v]et=on")

	tt.beforeView(func() { tt.ClickReporting(0) })
	t.Contains(tt.Trim(tt.Reporting()).String(), "FAIL")
	t.Contains(tt.Trim(tt.Reporting()).String(), "vet/src.go:11:26")
}

func (s *Report) Updated_package_vetted_if_vet_is_turned_on(t *T) {
	_, tt := s.fxSource(t, "vet")
	tt.ClickReporting(0)
	t.Not.Contains(tt.Trim(tt.Reporting()).String(), "fails if vetted")

	tt.ClickButton("vet=off")
	t.Contains(tt.ButtonBar().String(), "[v]et=on")
	t.Not.Contains(tt.Trim(tt.Reporting()).String(), "FAIL")

	tt.beforeWatch(func() { tt.golden.Touch("vet") })
	t.Contains(tt.Trim(tt.Reporting()).String(), "FAIL")
	t.Contains(tt.Trim(tt.Reporting()).String(), "vet/src.go:11:26")
}

func (s *Report) Current_package_not_vetted_if_vet_is_turned_off(t *T) {
	_, tt := s.fxSource(t, "vet")
	tt.beforeView(func() { tt.ClickButton("vet=off") })
	t.Contains(tt.Trim(tt.Reporting()).String(), "FAIL")

	tt.ClickButton("vet=on")
	tt.beforeWatch(func() { tt.golden.Touch("vet") })
	t.Not.Contains(tt.Trim(tt.Reporting()).String(), "FAIL")
}

func (s *Report) Race_in_current_package_if_race_is_turned_on(t *T) {
	_, tt := s.fxSource(t, "race")
	tt.ClickReporting(2) // unfold suite
	t.Contains(tt.Reporting().String(), "fails on race detector")
	t.Contains(tt.ButtonBar().String(), "[r]ace=off")

	tt.beforeView(func() { tt.ClickButton("race=off") })
	t.Contains(tt.ButtonBar().String(), "[r]ace=on")
	t.Contains(tt.Trim(tt.Reporting()).String(), "WARNING: DATA RACE")
}

func (s *Report) Race_in_selected_package_if_race_is_turned_on(t *T) {
	_, tt := s.fxSource(t, "race")
	tt.ClickReporting(0)
	t.Not.Contains(tt.Reporting().String(), "fails on race detector")

	tt.ClickButton("race=off")
	t.Contains(tt.ButtonBar().String(), "[r]ace=on")

	tt.beforeView(func() { tt.ClickReporting(0) })
	tt.beforeView(func() { tt.ClickReporting(2) })
	t.Contains(tt.Trim(tt.Reporting()).String(), "WARNING: DATA RACE")
}

func (s *Report) Race_in_updated_package_if_race_is_turned_on(t *T) {
	_, tt := s.fxSource(t, "race")
	tt.ClickReporting(0)
	t.Not.Contains(tt.Reporting().String(), "fails on race detector")

	tt.ClickButton("race=off")
	t.Contains(tt.ButtonBar().String(), "[r]ace=on")
	t.Not.Contains(tt.Trim(tt.Reporting()).String(), "WARNING: DATA RACE")

	tt.beforeWatch(func() { tt.golden.Touch("race") })
	// tt.beforeView(func() { tt.ClickReporting(2) })
	t.Contains(tt.Trim(tt.Reporting()).String(), "WARNING: DATA RACE")
}

func (s *Report) No_race_if_current_package_race_is_turned_off(t *T) {
	_, tt := s.fxSource(t, "race")
	tt.ClickReporting(2)
	tt.beforeView(func() { tt.ClickButton("race=off") })
	t.Contains(tt.Trim(tt.Reporting()).String(), "WARNING: DATA RACE")

	tt.ClickButton("race=on")
	tt.beforeWatch(func() { tt.golden.Touch("race") })
	t.Not.Contains(
		tt.Trim(tt.Reporting()).String(),
		"WARNING: DATA RACE",
	)
}

func TestReport(t *testing.T) {
	t.Parallel()
	Run(&Report{}, t)
}
