// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"fmt"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
	"github.com/slukits/gounit/pkg/lines/testdata/fx"
)

type AView struct {
	Suite
	fx FX
}

func (s *AView) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *AView) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, fx.New(t))
}

func (s *AView) TearDown(t *T) { s.fx.Del(t) }

func (s *AView) Sim_has_default_length(t *T) {
	rg := s.fx.Reg(t)
	rg.Resize(func(v *lines.View) {
		t.Eq(s.fx.DefaultLineCount, v.Len())
	})
	rg.Listen()
}

func (s *AView) Displays_an_error_if_len_to_small(t *T) {
	rg := s.fx.Reg(t)
	rg.Resize(func(v *lines.View) {
		v.SetMin(30)
	})
	rg.Listen()
}

func (s *AView) Adjust_length_according_to_a_resize_event(t *T) {
	rg, exp, resizeCount := s.fx.Reg(t, 1), 20, 0
	rg.Resize(func(v *lines.View) {
		switch resizeCount {
		case 0:
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 1:
			t.Eq(exp, v.Len())
		}
		resizeCount++
	})
	go rg.Listen()
	<-rg.NextEventProcessed // wait for initial resize to happen
	<-rg.SetNumberOfLines(exp)
	t.Eq(2, resizeCount)
}

func (s *AView) Adjusts_provided_screen_lines_at_resize_event(t *T) {
	rg, expFirst, expSecond, resizeCount := s.fx.Reg(t, 3), 15, 30, 0
	rg.Resize(func(v *lines.View) {
		switch resizeCount {
		case 0:
			got := 0
			v.LL().ForScreen(func(*lines.Line) { got++ })
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 1:
			got := 0
			v.LL().ForScreen(func(*lines.Line) { got++ })
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 2:
			got := 0
			v.LL().ForScreen(func(*lines.Line) { got++ })
			t.Eq(expFirst, got)
		case 3:
			got := 0
			v.LL().ForScreen(func(*lines.Line) { got++ })
			t.Eq(expSecond, got)
		}
		resizeCount++
	})
	go rg.Listen()
	<-rg.NextEventProcessed
	<-rg.SetNumberOfLines(s.fx.DefaultLineCount)
	<-rg.SetNumberOfLines(expFirst)
	<-rg.SetNumberOfLines(expSecond)
	t.Eq(4, resizeCount)
}

func (s *AView) Shows_an_error_if_resize_goes_below_min(t *T) {
	rg, first := s.fx.Reg(t, 1), true
	rg.Resize(func(v *lines.View) {
		if first {
			first = false
			v.SetMin(20)
		}
	})
	go rg.Listen()
	<-rg.NextEventProcessed
	rg.SetNumberOfLines(15) // since it errors it is not reported
	<-rg.Synced
	rg.QuitListening()
	t.Contains(rg.LastScreen, fmt.Sprintf(lines.ErrScreenFmt, 20))
}

func TestAView(t *testing.T) {
	t.Parallel()
	Run(&AView{}, t)
}

type DBG struct{ Suite }

func TestDBG(t *testing.T) { Run(&DBG{}, t) }
