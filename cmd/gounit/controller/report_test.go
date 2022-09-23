// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"testing"

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

	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
		fxExp["go/pass"]...,
	)
	t.StarMatched( // number of pkgs, suites, passed, failed
		tt.StatusBar().String(), "1", "2", "11", "0")
}

func (s *Report) Initially_suite_of_most_recent_test_file(t *T) {
	_, tt := s.fxSourceTouched(t, "mixed/pass", "mixed/pass/suite3_test.go")

	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
		fxExp["mixed/pass init"]...,
	)

	for _, s := range fxNotExp["mixed/pass"] {
		t.Not.Contains(tt.Reporting().String(), s)
	}
}

// func (s *Report) Folded_packages_on_reported_package_selection(t *T) {
// 	_, tt := s.fxSourceTouched(t, "mixed/pp", "mixed/pp/pkg0")
//
// 	t.StarMatched(
// 		tt.afterWatch(awReporting).String(),
// 		fxExp["mixed/pp/pkg0"]...,
// 	)
//
// 	tt.ClickReporting(0) // select package
// 	t.StarMatched(tt.Reporting().String(), fxExp["mixed/pp"]...)
// }
//
// func (s *Report) Selected_folded_package(t *T) {
// 	_, tt := s.fxSourceTouched(t, "mixed/pp", "mixed/pp/pkg0")
//
// 	t.StarMatched(
// 		tt.afterWatch(awReporting).String(),
// 		fxExp["mixed/pp/pkg0"]...,
// 	)
// 	tt.ClickReporting(0) // select package
// 	t.StarMatched(tt.Reporting().String(), fxExp["mixed/pp"]...)
// 	tt.ClickReporting(3) // select package 3
//
// 	t.SpaceMatched(
// 		tt.Reporting().String(),
// 		fxExp["mixed/pp/pkg3"]...,
// 	)
// }
//
// func (s *Report) Logged_text(t *T) {
// 	_, tt := s.fxSource(t, "logging")
//
// 	t.StarMatched(
// 		tt.afterWatch(awReporting).String(),
// 		fxExp["logging"]...,
// 	)
//
// 	tt.ClickReporting(2) // click on "go-tests"
// 	t.StarMatched(
// 		tt.Reporting().String(), fxExp["logging go-test"]...)
//
// 	// select the go-test-suite
// 	for i, l := range tt.Reporting() {
// 		if !strings.Contains(l.String(), "test go suite log") {
// 			continue
// 		}
// 		tt.ClickReporting(i)
// 		break
// 	}
// 	t.StarMatched(
// 		tt.Reporting().String(), fxExp["logging go-sub-test"]...)
//
// 	tt.ClickReporting(2) // back to folded view
// 	tt.ClickReporting(3) // select suite
// 	t.StarMatched(
// 		tt.Reporting().String(), fxExp["logging suite"]...)
// }
//
// func (s *Report) lineIsFailing(l lines.TestLine) bool {
// 	for i, r := range l.String() {
// 		if r == ' ' {
// 			continue
// 		}
// 		return l.Styles().Of(i).HasBG(tcell.ColorRed)
// 	}
// 	return false
// }
//
// func (s *Report) Failing_go_tests_ony_package(t *T) {
// 	failingLines := []int{0, 2, 7}
// 	_, tt := s.fxSource(t, "fail/gonly")
// 	got := tt.afterWatch(awReporting)
// 	t.FatalIfNot(t.True(len(got) > 7))
// 	for _, l := range failingLines {
// 		t.True(s.lineIsFailing(got[l]))
// 	}
//
// 	tt.ClickReporting(7)
// 	failingLines = []int{0, 2, 4}
// 	got = tt.Reporting()
// 	t.FatalIfNot(t.True(len(got) > 4))
// 	for _, l := range failingLines {
// 		t.True(s.lineIsFailing(got[l]))
// 	}
// }
//
// func (s *Report) Failing_package_due_to_compile_error(t *T) {
// 	_, tt := s.fxSource(t, "fail/compile")
// 	t.StarMatched(
// 		tt.afterWatch(awReporting).String(),
// 		fxExp["fail compile"]...,
// 	)
// }
//
// func (s *Report) Failing_package_s_failing_go_tests_initially(t *T) {
//
// }

func TestReport(t *testing.T) {
	t.Parallel()
	Run(&Report{}, t)
}
