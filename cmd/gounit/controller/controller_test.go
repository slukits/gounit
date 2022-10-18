// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"strings"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/cmd/gounit/model"
)

// Gounit tests the behavior of Controller.New which is identical with
// the behavior of main.
type Gounit struct {
	Suite
}

func (s *Gounit) Init(t *S) {
	initGolden(t)
}

func (s *Gounit) SetUp(t *T) { t.Parallel() }

func (s *Gounit) Fails_if_watching_fails(t *T) {
	fatale := false

	fxInit(t, InitFactories{
		Watcher: &watcherMock{watch: func() (
			<-chan *model.PackagesDiff, uint64, error,
		) {
			return nil, 0, errors.New("mock-err")
		}},
		Fatal: func(_ ...interface{}) { fatale = true },
	}, nil)

	t.True(fatale)
}

func (s *Gounit) fx(t *T) *Testing {
	return fx(t)
}

func (s *Gounit) Shows_initially_default_buttons(t *T) {
	exp := []string{"[s]witches", "[h]elp", "[a]bout", "[q]uit"}
	tt := s.fx(t)

	t.SpaceMatched(tt.ButtonBarCells(), exp...)
}

func (s *Gounit) Shows_initially_module_and_watched_pkg_name(t *T) {
	exp := []string{goldenModule, emptyPkg}
	tt := s.fx(t)

	t.StarMatched(tt.MessageBarCells(), exp...)
}

func (s *Gounit) Reports_initially_waiting_report(t *T) {
	tt := s.fx(t)
	t.Contains(tt.ReportCells(), initReport)
}

func (s *Gounit) Shows_help_screen(t *T) {
	tt := s.fx(t)
	tt.ClickButtons("help")
	got := tt.splitTrimmed(tt.ReportCells().Trimmed().String())
	t.SpaceMatched(help, got...)
}

func (s *Gounit) Shows_last_report_going_back_from_help(t *T) {
	tt := s.fx(t)
	exp := tt.ReportCells().Trimmed()
	tt.ClickButtons("help", "close")
	t.Eq(exp, tt.ReportCells().Trimmed())
}

func (s *Gounit) Shows_about_screen(t *T) {
	tt := s.fx(t)
	tt.ClickButtons("about")
	got := tt.splitTrimmed(tt.ReportCells().Trimmed().String())
	t.SpaceMatched(about, got...)
}

func (s *Gounit) Shows_last_report_going_back_from_about(t *T) {
	tt := s.fx(t)
	exp := tt.ReportCells().Trimmed()
	tt.ClickButtons("about", "close")
	t.Eq(exp, tt.ReportCells().Trimmed())
}

func (s *Gounit) fxSource(t *T, dir string) *Testing {
	return fxSource(t, dir)
}

func (s *Gounit) Folds_and_unfolds_go_tests_only_package(t *T) {
	tt := s.fxSource(t, "go/pass")
	t.StarMatched(tt.ReportCells(), fxExp["go/pass"]...)

	tt.ClickReporting(0)
	t.Not.StarMatched(tt.ReportCells(), fxExp["go/pass"]...)
	t.Contains(tt.ReportCells(), "go/pass")

	tt.ClickReporting(0)
	t.StarMatched(tt.ReportCells(), fxExp["go/pass"]...)
}

func (s *Gounit) Folds_selected_unfolded_suite(t *T) {
	tt := s.fxSource(t, "mixed/pass")
	tt.ClickReporting(3) // select suite 1
	t.Contains(tt.ReportCells(), "suite test 1 1")

	tt.ClickReporting(2)
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass fold suite"]...)
}

func (s *Gounit) Unfolds_selected_suite(t *T) {
	tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass fold suite"]...)

	tt.ClickReporting(4)
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass unfold suite"]...)
}

func (s *Gounit) Unfolds_go_tests_with_folded_go_suites(t *T) {
	tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass fold suite"]...)

	tt.ClickReporting(2)
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass go folded subs"]...)
}

func (s *Gounit) Unfolds_folded_go_suite_in_go_tests(t *T) {
	tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass fold suite"]...)

	tt.ClickReporting(2) // unfold go-tests
	tt.ClickReporting(8) // unfold go-suite
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass go unfolded suite"]...)
}

func (s *Gounit) Folds_go_suite_on_unfolded_go_suite_in_go_tests(t *T) {
	tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass fold suite"]...)
	tt.ClickReporting(2) // unfold go-tests
	tt.ClickReporting(8) // unfold go-suite
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass go unfolded suite"]...)

	tt.ClickReporting(4)
	t.StarMatched(
		tt.ReportCells(), fxExp["mixed/pass go folded subs"]...)
}

func (s *Gounit) fxSourceTouched(t *T, dir, touch string) *Testing {
	return fxSourceTouched(t, dir, touch)
}

func (s *Gounit) Folds_package_on_unfolded_package_selection(t *T) {
	tt := s.fxSourceTouched(t, "mixed/pp", "mixed/pp/pkg0")

	t.StarMatched(tt.ReportCells(), fxExp["mixed/pp/pkg0"]...)
	t.Not.SpaceMatched(tt.ReportCells(), fxExp["mixed/pp"]...)

	tt.ClickReporting(0) // select package
	t.StarMatched(tt.ReportCells(), fxExp["mixed/pp"]...)
}

func (s *Gounit) Unfolds_package_on_folded_package_selection(t *T) {
	tt := s.fxSourceTouched(t, "mixed/pp", "mixed/pp/pkg0")

	t.StarMatched(tt.ReportCells(), fxExp["mixed/pp/pkg0"]...)

	tt.ClickReporting(0) // select package
	t.StarMatched(tt.ReportCells(), fxExp["mixed/pp"]...)
	tt.ClickReporting(3) // select package 3

	t.SpaceMatched(tt.ReportCells(), fxExp["mixed/pp/pkg3"]...)
}

func (s *Gounit) Locks_selected_suite_on_test_file_updated(t *T) {
	tt := s.fxSource(t, "twosuites")
	t.FatalIfNot(t.Contains(tt.ReportCells(), "suite 1"))
	tt.ClickReporting(3) // select suite 1
	t.FatalIfNot(t.Contains(tt.ReportCells(), "suite 1"))

	tt.before(func() {
		tt.golden.Touch("twosuites/pass_test.go")
	})
	t.Contains(tt.ReportCells(), "suite 1")
}

func (s *Gounit) Locks_selected_go_tests_on_test_file_update(t *T) {
	tt := s.fxSource(t, "twosuites")
	t.FatalIfNot(t.Contains(tt.ReportCells(), "go-tests"))
	tt.ClickReporting(2) // select go-tests
	t.FatalIfNot(t.Contains(tt.ReportCells(), "go-tests"))

	tt.before(func() {
		tt.golden.Touch("twosuites/pass_test.go")
	})
	t.Contains(tt.ReportCells(), "go-tests")
}

func (s *Gounit) Locks_selected_go_suite_on_test_file_update(t *T) {
	tt := s.fxSource(t, "twosuites")
	t.FatalIfNot(t.Contains(tt.ReportCells(), "go-tests"))
	tt.ClickReporting(2) // select go-tests
	tt.ClickReporting(8) // select go-suite
	t.FatalIfNot(t.Contains(tt.ReportCells(), "p4 sub 3"))

	tt.before(func() {
		tt.golden.Touch("twosuites/pass_test.go")
	})
	t.Contains(tt.ReportCells(), "p4 sub 3")
}

func (s *Gounit) Stops_reporting_a_removed_package(t *T) {
	tt := s.fxSource(t, "del")
	tt.ClickReporting(0)
	t.StarMatched(tt.ReportCells(), fxExp["del before"]...)

	tt.before(func() { tt.golden.Rm("del/pkg1") })
	t.Not.Contains(tt.ReportCells(), "del/pkg1")
	t.Contains(tt.ReportCells(), "del/pkg2")
}

func (s *Gounit) Reports_source_stats_if_stats_turned_on(t *T) {
	tt := s.fxSource(t, "srcstats")
	tt.ClickButton("switches")
	t.Contains(tt.ButtonBarCells(), "[s]tats=off")
	t.Not.Contains(tt.StatusBarCells(), "source-stats")
	tt.beforeView(func() { tt.ClickButton("stats=off") })

	t.Contains(tt.ButtonBarCells(), "[s]tats=on")
	t.Contains(tt.StatusBarCells(), "source-stats")
	t.Contains(tt.ReportCells(), "5/2 9/3/26")
}

func (s *Gounit) Shows_reported_package_ID_in_message_bar(t *T) {
	tt := s.fxSource(t, "mixed/pp")
	exp := ""
	ss := []string{} // extract reported package name
	for _, r := range strings.TrimSpace(tt.ReportCells()[0].String()) {
		if r == ' ' {
			break
		}
		ss = append(ss, string(r))
	}
	exp = strings.Join(ss, "")

	t.Contains(tt.MessageBarCells(), exp)
}

func (s *Gounit) Suspends_model_change_reporting_showing_something_else(
	t *T,
) {
	tt := s.fxSource(t, "mixed/pp")
	tt.ClickButtons("about")
	t.FatalIfNot(t.Contains(tt.ReportCells(), "gounit Copyright"))

	tt.beforeWatch(func() { tt.golden.Touch("mixed/pp/pkg3") })
	t.Contains(tt.ReportCells(), "gounit Copyright")
}

func (s *Gounit) Sticks_to_sole_failed_suite_on_passing(t *T) {
	tt := s.fxSource(t, "fail/suite")
	t.StarMatched(tt.ReportCells(), "suite test 2", "fail_test")
	dir := tt.golden.Child("fail").Child("suite")

	tt.beforeWatch(func() {
		dir.WriteContent("fail_test.go", fxPassingSuite)
	})
	got := tt.ReportCells()
	t.Contains(got, "suite test 2")
	t.Not.Contains(got, "fail_test")
}

func (s *Gounit) Sticks_to_sole_failed_go_suite_on_passing(t *T) {
	tt := s.fxSource(t, "fail/gosuite")
	t.StarMatched(tt.ReportCells(), "p2 sub", "fail_test")
	dir := tt.golden.Child("fail").Child("gosuite")

	tt.beforeWatch(func() {
		dir.WriteContent("fail_test.go", fxPassingGoSuite)
	})
	got := tt.ReportCells()
	t.Contains(got, "p2 sub")
	t.Not.Contains(got, "fail_test")
}

func (s *Gounit) Folds_failing_suite(t *T) {
	tt := s.fxSource(t, "fail/suite")
	t.Contains(tt.ReportCells(), "fail_test")

	tt.ClickReporting(2)
	t.Not.Contains(tt.ReportCells(), "fail_test")
}

func (s *Gounit) Folds_failing_go_suite(t *T) {
	tt := s.fxSource(t, "fail/gosuite")
	t.Contains(tt.ReportCells(), "fail_test")

	tt.ClickReporting(2)
	t.Not.Contains(tt.ReportCells(), "fail_test")
}

func TestGounit(t *testing.T) {
	t.Parallel()
	Run(&Gounit{}, t)
}
