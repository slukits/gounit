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
	})

	t.True(fatale)
}

func (s *Gounit) Listens_to_events_if_not_fatale(t *T) {
	ee, _ := fx(t, s)
	t.True(ee.IsListening())
}

func (s *Gounit) fx(t *T) (*lines.Events, *Testing) {
	return fx(t, s)
}

func (s *Gounit) Shows_initially_default_buttons(t *T) {
	exp := []string{"[p]kgs", "[s]uites=off", "[a]rgs", "[m]ore"}
	_, tt := s.fx(t)

	t.SpaceMatched(tt.ButtonBar().String(), exp...)
}

func (s *Gounit) Shows_initially_module_and_watched_pkg_name(t *T) {
	exp := []string{mckModule, mckSourceDir}
	_, tt := s.fx(t)

	t.StarMatched(tt.MessageBar().String(), exp...)
}

func (s *Gounit) Shows_initially_initial_report(t *T) {
	_, tt := s.fx(t)
	t.Contains(tt.Reporting().String(), initReport)
}

func (s *Gounit) Shows_help_screen(t *T) {
	_, tt := s.fx(t)
	tt.ClickButtons("more", "help")
	got := tt.SplitTrimmed(tt.Trim(tt.Reporting()).String())
	t.SpaceMatched(help, got...)
}

func (s *Gounit) Shows_last_report_going_back_from_help(t *T) {
	_, tt := s.fx(t)
	exp := tt.Trim(tt.Reporting()).String()
	tt.ClickButtons("more", "help", "back")
	t.Eq(exp, tt.Trim(tt.Reporting()).String())
}

func (s *Gounit) Shows_about_screen(t *T) {
	_, tt := s.fx(t)
	tt.ClickButtons("more", "about")
	got := tt.SplitTrimmed(tt.Trim(tt.Reporting()).String())
	t.SpaceMatched(about, got...)
}

func (s *Gounit) Shows_last_report_going_back_from_about(t *T) {
	_, tt := s.fx(t)
	exp := tt.Trim(tt.Reporting()).String()
	tt.ClickButtons("more", "about", "back")
	t.Eq(exp, tt.Trim(tt.Reporting()).String())
}

func TestGounit(t *testing.T) {
	t.Parallel()
	Run(&Gounit{}, t)
}
