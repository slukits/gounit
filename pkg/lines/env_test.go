// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"fmt"
	"testing"

	. "github.com/slukits/gounit"
)

type AEnv struct {
	Suite
	fx *FX
}

func (s *AEnv) Init(t *I) {
	s.fx = NewFX()
}

func (s *AEnv) SetUp(t *T) {
	t.Parallel()
	s.fx.New(t)
}

func (s *AEnv) TearDown(t *T) { s.fx.Del(t) }

func (s *AEnv) For_testing_has_default_length(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(e *Env) {
		t.Eq(s.fx.DefaultLineCount, e.Len())
	})
	ee.Listen()
}

func (s *AEnv) Displays_an_error_if_len_to_small(t *T) {
	ee, tt := s.fx.For(t)
	ee.Resize(func(e *Env) { e.SetMin(30) })
	ee.Listen()
	t.Contains(fmt.Sprint(tt.LastScreen), fmt.Sprintf(ErrScreenFmt, 30))
}

func (s *AEnv) Adjust_length_according_to_a_resize_event(t *T) {
	ee, tt := s.fx.For(t, 2)
	exp, resizeCount := 20, 0
	ee.Resize(func(e *Env) {
		switch resizeCount {
		case 0:
			t.Eq(s.fx.DefaultLineCount, e.Len())
		case 1:
			t.Eq(exp, e.Len())
		}
		resizeCount++
	})
	tt.FireResize(exp)
	t.Eq(2, resizeCount)
}

func (s *AEnv) Adjusts_provided_screen_lines_at_resize_event(t *T) {
	ee, tt := s.fx.For(t, 4)
	expFirst, expSecond, resizeCount := 15, 30, 0
	ee.Resize(func(e *Env) {
		got := 0
		switch resizeCount {
		case 0:
			e.LL().ForScreen(func(*Line) { got++ })
			t.Eq(s.fx.DefaultLineCount, e.Len())
		case 1:
			e.LL().ForScreen(func(*Line) { got++ })
			t.Eq(s.fx.DefaultLineCount, e.Len())
		case 2:
			e.LL().ForScreen(func(*Line) { got++ })
			t.Eq(expFirst, got)
		case 3:
			e.LL().ForScreen(func(*Line) { got++ })
			t.Eq(expSecond, got)
		}
		resizeCount++
	})
	tt.FireResize(s.fx.DefaultLineCount)
	tt.FireResize(expFirst)
	tt.FireResize(expSecond)
	t.Eq(4, resizeCount)
}

func (s *AEnv) Shows_an_error_if_resize_goes_below_min(t *T) {
	ee, tt := s.fx.For(t, 2)
	first := true
	ee.Resize(func(e *Env) {
		if first {
			first = false
			e.SetMin(20)
		}
	})
	ee.Listen()
	t.False(first)
	tt.FireResize(15) // not reported
	t.True(ee.IsListening())
	ee.QuitListening()
	t.Contains(fmt.Sprintf(tt.LastScreen), fmt.Sprintf(ErrScreenFmt, 20))
}

func TestAEnv(t *testing.T) {
	t.Parallel()
	Run(&AEnv{}, t)
}
