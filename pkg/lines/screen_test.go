// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"fmt"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

type AScreen struct {
	Suite
	fx FX
}

func (s *AScreen) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *AScreen) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, New(t))
}

func (s *AScreen) TearDown(t *T) { s.fx.Del(t) }

func (s *AScreen) Sim_has_default_length(t *T) {
	ee := s.fx.EE(t)
	ee.Resize(func(v *lines.Screen) {
		t.Eq(s.fx.DefaultLineCount, v.Len())
	})
	ee.Listen()
}

func (s *AScreen) Displays_an_error_if_len_to_small(t *T) {
	ee := s.fx.EE(t)
	ee.Resize(func(v *lines.Screen) {
		v.SetMin(30)
	})
	ee.Listen()
}

func (s *AScreen) Adjust_length_according_to_a_resize_event(t *T) {
	ee, exp, resizeCount := s.fx.EE(t, 1), 20, 0
	ee.Resize(func(v *lines.Screen) {
		switch resizeCount {
		case 0:
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 1:
			t.Eq(exp, v.Len())
		}
		resizeCount++
	})
	ee.SetNumberOfLines(exp)
	t.Eq(2, resizeCount)
	t.False(ee.IsPolling())
}

func (s *AScreen) Adjusts_provided_screen_lines_at_resize_event(t *T) {
	ee, expFirst, expSecond, resizeCount := s.fx.EE(t, 3), 15, 30, 0
	ee.Resize(func(v *lines.Screen) {
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
	ee.SetNumberOfLines(s.fx.DefaultLineCount)
	ee.SetNumberOfLines(expFirst)
	ee.SetNumberOfLines(expSecond)
	t.Eq(4, resizeCount)
	t.False(ee.IsPolling())
}

func (s *AScreen) Shows_an_error_if_resize_goes_below_min(t *T) {
	ee, first := s.fx.EE(t, 1), true
	ee.Resize(func(v *lines.Screen) {
		if first {
			first = false
			v.SetMin(20)
		}
	})
	ee.Listen()
	t.False(first)
	ee.SetNumberOfLines(15) // not reported
	t.True(ee.IsPolling())
	ee.QuitListening()
	t.Contains(ee.LastScreen, fmt.Sprintf(lines.ErrScreenFmt, 20))
}

func TestAView(t *testing.T) {
	t.Parallel()
	Run(&AScreen{}, t)
}
