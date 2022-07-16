// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

type Lines struct {
	Suite
	fx FX
}

func (s *Lines) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *Lines) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, New(t))
}

func (s *Lines) TearDown(t *T) { s.fx.Del(t) }

func (s *Lines) Has_initially_view_len_many_lines(t *T) {
	ee := s.fx.EE(t)
	ee.Resize(func(v *lines.Env) {
		t.Eq(v.Len(), v.LL().Len())
		got := 0
		v.LL().For(func(l *lines.Line) { got++ })
		t.Eq(v.Len(), got)
	})
	ee.Listen()
}

func (s *Lines) First_screen_line_defaults_to_zero(t *T) {
	ee := s.fx.EE(t)
	ee.Resize(func(v *lines.Env) {
		t.Eq(0, v.LL().FirstScreenLine())
	})
	ee.Listen()
}

func (s *Lines) Ignores_first_screen_line_update_if_out_of_bound(t *T) {
	ee := s.fx.EE(t)
	ee.Resize(func(v *lines.Env) {
		t.Eq(5, v.LL().SetFirstScreenLine(5).FirstScreenLine())
		t.Eq(5, v.LL().SetFirstScreenLine(-1).FirstScreenLine())
		t.Eq(5, v.LL().SetFirstScreenLine(42).FirstScreenLine())
	})
	ee.Listen()
}

func (s *Lines) Provides_all_screen_lines_from_the_first_screen_line(
	t *T,
) {
	ee := s.fx.EE(t)
	ee.Resize(func(v *lines.Env) {
		exp := 0
		v.LL().ForScreen(func(l *lines.Line) {
			t.Eq(exp, l.Idx)
			exp++
		})
		v.LL().SetFirstScreenLine(20)
		exp = 20
		v.LL().ForScreen(func(l *lines.Line) {
			t.Eq(exp, l.Idx)
			exp++
		})
	})
	ee.Listen()
}

func (s *Lines) Provides_zero_line_for_out_of_bound_requests(t *T) {
	ee := s.fx.EE(t)
	ee.Resize(func(v *lines.Env) {
		t.True(lines.Zero == v.LL().Line(s.fx.DefaultLineCount))
	})
	ee.Listen()
}

func (s *Lines) Are_Increased_as_requested(t *T) {
	ee, exp := s.fx.EE(t), 42
	ee.Resize(func(v *lines.Env) {
		got := 0
		v.LL().ForN(-1, func(l *lines.Line) { got++ })
		t.Eq(0, got)
		v.LL().ForN(exp, func(l *lines.Line) { got++ })
		t.Eq(exp, got)
	})
	ee.Listen()
	t.False(ee.IsPolling())
}

func TestLines(t *testing.T) {
	t.Parallel()
	Run(&Lines{}, t)
}
