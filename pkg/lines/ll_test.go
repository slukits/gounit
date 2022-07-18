// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/slukits/gounit"
)

type lines struct {
	Suite
	fx *FX
}

func (s *lines) Init(t *I) {
	s.fx = NewFX()
	s.fx.DefaultLineCount = 25
}

func (s *lines) SetUp(t *T) {
	t.Parallel()
	s.fx.New(t)
}

func (s *lines) TearDown(t *T) { s.fx.Del(t) }

func (s *lines) Has_initially_env_len_many_lines(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(e *Env) {
		t.Eq(e.Len(), e.LL().Len())
		got := 0
		e.LL().For(func(l *Line) { got++ })
		t.Eq(e.Len(), got)
	})
	ee.Listen()
}

func (s *lines) First_screen_line_defaults_to_zero(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(e *Env) {
		t.Eq(0, e.LL().FirstScreenLine())
	})
	ee.Listen()
}

func (s *lines) Ignores_first_screen_line_update_if_out_of_bound(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(e *Env) {
		t.Eq(5, e.LL().SetFirstScreenLine(5).FirstScreenLine())
		t.Eq(5, e.LL().SetFirstScreenLine(-1).FirstScreenLine())
		t.Eq(5, e.LL().SetFirstScreenLine(42).FirstScreenLine())
	})
	ee.Listen()
}

func (s *lines) Provides_all_screen_lines_from_the_first_screen_line(
	t *T,
) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(e *Env) {
		exp := 0
		e.LL().ForScreen(func(l *Line) {
			t.Eq(exp, l.Idx)
			exp++
		})
		e.LL().SetFirstScreenLine(20)
		exp = 20
		e.LL().ForScreen(func(l *Line) {
			t.Eq(exp, l.Idx)
			exp++
		})
	})
	ee.Listen()
}

func (s *lines) Provides_zero_line_for_out_of_bound_requests(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(e *Env) {
		t.True(Zero == e.LL().Line(s.fx.DefaultLineCount))
	})
	ee.Listen()
}

func (s *lines) Are_Increased_as_requested(t *T) {
	ee, _ := s.fx.For(t)
	exp := 42
	ee.Resize(func(e *Env) {
		got := 0
		e.LL().ForN(-1, func(l *Line) { got++ })
		t.Eq(0, got)
		e.LL().ForN(exp, func(l *Line) { got++ })
		t.Eq(exp, got)
	})
	ee.Listen()
}

func (s *lines) Prints_changed_content_of_screen_lines_to_screen(t *T) {
	ee, tt := s.fx.For(t)
	exp := &strings.Builder{}
	ee.Resize(func(e *Env) {
		for i := 0; i < e.Len(); i++ {
			_, err := fmt.Fprintf(exp, "line%d\n", i)
			t.FatalOn(err)
			e.LL().Line(i).Set(fmt.Sprintf("line%d", i))
		}
	})
	ee.Listen()
	t.Eq(strings.TrimSpace(exp.String()), fmt.Sprint(tt.LastScreen))
}

func TestLines(t *testing.T) {
	t.Parallel()
	Run(&lines{}, t)
}
