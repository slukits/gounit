// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/lines"
)

// Gounit tests the behavior of Controller.New which is identical with
// the behavior of main.
type Gounit struct {
	Suite
	Fixtures
}

func (s *Gounit) Init(t *S) {
	initGolden(t)
}

func (s *Gounit) SetUp(t *T) { t.Parallel() }

func (s *Gounit) TearDown(t *T) {
	fx := s.Get(t)
	if fx == nil {
		return
	}
	fx.(func())()
}

func (s *Gounit) Fails_if_watching_fails(t *T) {
	fatale := false

	fxInit(t, s, InitFactories{
		Watcher: &watcherMock{watch: func() (
			<-chan *model.PackagesDiff, uint64, error,
		) {
			return nil, 0, errors.New("mock-err")
		}},
		Fatal: func(_ ...interface{}) { fatale = true },
	}, nil)

	t.True(fatale)
}

func (s *Gounit) Listens_to_events_if_not_fatale(t *T) {
	ee, _ := fx(t, s)
	t.True(ee.IsListening())
}

func (s *Gounit) fx(t *T) (*lines.Events, *Testing) {
	return fx(t, s)
}

func (s *Gounit) fxSource(t *T, dir string) (*lines.Events, *Testing) {
	return fxSource(t, s, dir)
}

func (s *Gounit) fxSourceTouched(
	t *T, dir, touch string,
) (*lines.Events, *Testing) {

	return fxSourceTouched(t, s, dir, touch)
}

func (s *Gounit) Shows_initially_default_buttons(t *T) {
	exp := []string{"[v]et=off", "[r]ace=off", "[s]tats=off", "[m]ore"}
	_, tt := s.fx(t)

	t.SpaceMatched(tt.ButtonBar().String(), exp...)
}

func (s *Gounit) Shows_initially_module_and_watched_pkg_name(t *T) {
	exp := []string{goldenModule, emptyPkg}
	_, tt := s.fx(t)

	t.StarMatched(tt.MessageBar().String(), exp...)
}

func (s *Gounit) Shows_initially_initial_report(t *T) {
	_, tt := s.fx(t)
	t.Contains(tt.Reporting().String(), initReport)
}

func (s *Gounit) Shows_help_screen(t *T) {
	_, tt := s.fx(t)
	tt.clickButtons("more", "help")
	got := tt.splitTrimmed(tt.Trim(tt.Reporting()).String())
	t.SpaceMatched(help, got...)
}

func (s *Gounit) Shows_last_report_going_back_from_help(t *T) {
	_, tt := s.fx(t)
	exp := tt.Trim(tt.Reporting()).String()
	tt.clickButtons("more", "help", "back")

	t.Eq(exp, tt.Trim(tt.Reporting()).String())
}

func (s *Gounit) Shows_about_screen(t *T) {
	_, tt := s.fx(t)
	tt.clickButtons("more", "about")
	got := tt.splitTrimmed(tt.Trim(tt.Reporting()).String())
	t.SpaceMatched(about, got...)
}

func (s *Gounit) Shows_last_report_going_back_from_about(t *T) {
	_, tt := s.fx(t)
	exp := tt.Trim(tt.Reporting()).String()
	tt.clickButtons("more", "about", "back")
	t.Eq(exp, tt.Trim(tt.Reporting()).String())
}

func (s *Gounit) Shows_last_report_going_back_from_about_and_help(t *T) {
	_, tt := s.fx(t)
	exp := tt.Trim(tt.Reporting()).String()
	tt.clickButtons("more", "about", "help", "back")
	t.Eq(exp, tt.Trim(tt.Reporting()).String())
}

func (s *Gounit) Folds_and_unfolds_go_tests_only_package(t *T) {
	_, tt := s.fxSource(t, "go/pass")
	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
		fxExp["go/pass"]...,
	)

	tt.ClickReporting(0)
	t.Not.StarMatched(tt.Reporting().String(), fxExp["go/pass"]...)
	t.Contains(tt.Reporting().String(), "go/pass")

	tt.ClickReporting(0)
	t.StarMatched(tt.Reporting().String(), fxExp["go/pass"]...)
}

func (s *Gounit) Folds_selected_unfolded_suite(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")
	tt.afterWatch(func() { tt.ClickReporting(3) }) // select suite 1
	t.Contains(tt.Reporting().String(), "suite test 1 1")

	tt.ClickReporting(2)
	t.StarMatched(
		tt.Reporting().String(), fxExp["mixed/pass fold suite"]...)
}

func (s *Gounit) Unfolds_selected_suite(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
		fxExp["mixed/pass fold suite"]...,
	)

	tt.ClickReporting(4)
	t.StarMatched(
		tt.Reporting().String(), fxExp["mixed/pass unfold suite"]...)
}

func (s *Gounit) Unfolds_go_tests_with_folded_go_suites(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
		fxExp["mixed/pass fold suite"]...,
	)

	tt.ClickReporting(2)
	t.StarMatched(
		tt.Reporting().String(), fxExp["mixed/pass go folded subs"]...)
}

func (s *Gounit) Unfolds_folded_go_suite_in_go_tests(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
		fxExp["mixed/pass fold suite"]...,
	)

	tt.ClickReporting(2) // unfold go-tests
	tt.ClickReporting(8) // unfold go-suite
	t.StarMatched(
		tt.Reporting().String(), fxExp["mixed/pass go unfolded suite"]...)
}

func (s *Gounit) Folds_go_suite_on_unfolded_go_suite_in_go_tests(t *T) {
	_, tt := s.fxSource(t, "mixed/pass")
	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
		fxExp["mixed/pass fold suite"]...,
	)
	tt.ClickReporting(2) // unfold go-tests
	tt.ClickReporting(8) // unfold go-suite
	t.StarMatched(
		tt.Reporting().String(), fxExp["mixed/pass go unfolded suite"]...)

	tt.ClickReporting(3)
	t.StarMatched(
		tt.Reporting().String(), fxExp["mixed/pass go folded subs"]...)
}

func (s *Gounit) Folds_package_on_unfolded_package_selection(t *T) {
	_, tt := s.fxSourceTouched(t, "mixed/pp", "mixed/pp/pkg0")

	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
		fxExp["mixed/pp/pkg0"]...,
	)
	t.Not.SpaceMatched(tt.Reporting().String(), fxExp["mixed/pp"]...)

	tt.ClickReporting(0) // select package
	t.StarMatched(tt.Reporting().String(), fxExp["mixed/pp"]...)
}

func (s *Gounit) Unfolds_package_on_folded_package_selection(t *T) {
	_, tt := s.fxSourceTouched(t, "mixed/pp", "mixed/pp/pkg0")

	t.StarMatched(
		tt.afterWatchScr(awReporting).String(),
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

func (s *Gounit) Locks_selected_suite_on_test_file_updated(t *T) {
	_, tt := s.fxSource(t, "twosuites")
	tt.afterWatch(func() {
		t.FatalIfNot(t.Contains(tt.Reporting().String(), "suite 1"))
	})
	tt.ClickReporting(3) // select suite 1
	t.FatalIfNot(t.Contains(tt.Reporting().String(), "suite 1"))

	tt.beforeWatch(func() {
		tt.golden.Touch("twosuites/pass_test.go")
	})
	t.Contains(tt.Reporting().String(), "suite 1")
}

func (s *Gounit) Locks_selected_go_tests_on_test_file_update(t *T) {
	_, tt := s.fxSource(t, "twosuites")
	tt.afterWatch(func() {
		t.FatalIfNot(t.Contains(tt.Reporting().String(), "go-tests"))
	})
	tt.ClickReporting(2) // select go-tests
	t.FatalIfNot(t.Contains(tt.Reporting().String(), "go-tests"))

	tt.beforeWatch(func() {
		tt.golden.Touch("twosuites/pass_test.go")
	})
	t.Contains(tt.Reporting().String(), "go-tests")
}

func (s *Gounit) Locks_selected_go_suite_on_test_file_update(t *T) {
	_, tt := s.fxSource(t, "twosuites")
	tt.afterWatch(func() {
		t.FatalIfNot(t.Contains(tt.Reporting().String(), "go-tests"))
	})
	tt.ClickReporting(2) // select go-tests
	tt.ClickReporting(8) // select go-suite
	t.FatalIfNot(t.Contains(tt.Reporting().String(), "p4 sub 3"))

	tt.beforeWatch(func() {
		tt.golden.Touch("twosuites/pass_test.go")
	})
	t.Contains(tt.Reporting().String(), "p4 sub 3")
}

func (s *Gounit) Stops_reporting_a_removed_package(t *T) {
	_, tt := s.fxSource(t, "del")
	tt.afterWatch(func() {
		tt.ClickReporting(0)
	})
	t.StarMatched(tt.Reporting().String(), fxExp["del before"]...)

	tt.beforeWatch(func() {
		tt.golden.Rm("del/pkg1")
	})
	t.Not.Contains(tt.Reporting().String(), "del/pkg1")
	t.Contains(tt.Reporting().String(), "del/pkg2")
	// that shouldn't be needed because before watch shouldn't return
	// before the view is updated but somehow it does.
	// t.Within((&TimeStepper{}), func() bool {
	// 	return strings.Contains(tt.Reporting().String(), "del/pkg2") &&
	// 		!strings.Contains(tt.Reporting().String(), "del/pkg1")
	// })
}

func (s *Gounit) Reports_source_stats_if_stats_turned_on(t *T) {
	_, tt := s.fxSource(t, "srcstats")
	tt.afterWatch(func() {
		t.Contains(tt.ButtonBar().String(), "[s]tats=off")
		t.Not.Contains(tt.StatusBar().String(), "source-stats")
		tt.ClickButton("stats=off")
	})

	t.Contains(tt.ButtonBar().String(), "[s]tats=on")
	t.Contains(tt.StatusBar().String(), "source-stats")
	t.Contains(tt.Reporting().String(), "5/2 9/3/26")
}

func TestGounit(t *testing.T) {
	t.Parallel()
	Run(&Gounit{}, t)
}
