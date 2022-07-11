// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"testing"
	"time"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
	"github.com/slukits/gounit/pkg/lines/testdata/mck"
)

type NewView struct{ Suite }

func (s *NewView) Fails_if_tcell_s_screen_creation_fails(t *T) {
	lines.SetScreenFactory(&mck.ScreenFactory{Fail: true})
	_, err := lines.NewView()
	t.ErrIs(err, mck.ScreenErr)
}

func (s *NewView) Fails_if_tcell_s_screen_init_fails(t *T) {
	lines.SetScreenFactory(&mck.ScreenFactory{FailInit: true})
	_, err := lines.NewView()
	t.ErrIs(err, mck.InitErr)
}

func (s *NewView) Succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(&mck.ScreenFactory{})
	_, err := lines.NewView()
	t.FatalOn(err)
}

func (s *NewView) May_fail_in_graphical_test_environment(t *T) {
	// sole purpose of this test is keeping coverage at 100%
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	v, err := lines.NewView()
	if err == nil {
		lines.ExtractLib(v).Fini()
	}
}

func (s *NewView) Sim_fails_if_tcell_s_sim_init_fails(t *T) {
	lines.SetScreenFactory(&mck.ScreenFactory{FailInit: true})
	_, _, err := lines.NewSim()
	t.ErrIs(err, mck.InitErr)
}

func (s *NewView) Sim_succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	_, lib, err := lines.NewSim()
	t.FatalOn(err)
	lib.Fini()
}

func (s *NewView) Finalize(t *F) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
}

// TestView can not run in parallel since its tests manipulate the
// package-global state which is necessary to mock errors of the
// tcell-library.
func TestNewView(t *testing.T) {
	Run(&NewView{}, t)
}

type View struct {
	Suite
	fx FX
}

type FX struct {
	*Fixtures
	DefaultLineCount int
}

func (f *FX) View(t *T, maxEvt ...int) *mck.View {
	if len(maxEvt) == 0 {
		return f.Get(t).(*mck.View)
	}
	v := f.Get(t).(*mck.View)
	v.MaxEvents = maxEvt[0]
	return v
}

func (s *View) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *View) SetUp(t *T) { s.fx.Set(t, mck.NewView(t)) }

func (s *View) Provides_initial_resize_event(t *T) {
	v, resizeListenerCalled := s.fx.View(t), false
	v.Listeners.Resize = func(v *lines.View) {
		resizeListenerCalled = true
	}
	v.Listen()
	t.True(resizeListenerCalled)
}

func (s *View) Sim_provides_default_length(t *T) {
	v := s.fx.View(t)
	v.Listeners.Resize = func(v *lines.View) {
		t.Eq(s.fx.DefaultLineCount, v.Len())
	}
	v.Listen()
}

func (s *View) Provides_len_many_lines(t *T) {
	v, got := s.fx.View(t), 0
	v.Listeners.Resize = func(v *lines.View) {
		v.For(func(*lines.Line) { got++ })
	}
	v.Listen()
	t.Eq(s.fx.DefaultLineCount, got)
}

func (s *View) Resize_adjust_length_accordingly(t *T) {
	v, exp, resizeCount := s.fx.View(t, 1), 20, 0
	v.Listeners.Resize = func(v *lines.View) {
		switch resizeCount {
		case 0:
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 1:
			t.Eq(exp, v.Len())
		}
		resizeCount++
	}
	go v.Listen()
	<-v.NextEventProcessed // wait for initial resize to happen
	v.SetNumberOfLines(exp)
	<-v.NextEventProcessed // wait for resize event to happen
	t.Eq(2, resizeCount)
}

func (s *View) Resize_adjusts_the_provided_lines(t *T) {
	v, expFirst, expSecond, resizeCount := s.fx.View(t, 3), 15, 20, 0
	v.Listeners.Resize = func(v *lines.View) {
		switch resizeCount {
		case 0:
			got := 0
			v.For(func(*lines.Line) { got++ })
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 1:
			got := 0
			v.For(func(*lines.Line) { got++ })
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 2:
			got := 0
			v.For(func(*lines.Line) { got++ })
			t.Eq(expFirst, got)
		case 3:
			got := 0
			v.For(func(*lines.Line) { got++ })
			t.Eq(expSecond, got)
		}
		resizeCount++
	}
	go v.Listen()
	<-v.NextEventProcessed
	v.SetNumberOfLines(s.fx.DefaultLineCount)
	<-v.NextEventProcessed
	v.SetNumberOfLines(expFirst)
	<-v.NextEventProcessed
	v.SetNumberOfLines(expSecond)
	<-v.NextEventProcessed
	t.Eq(4, resizeCount)
}

func (s *View) Unregistered_rune_events_are_ignored(t *T) {
	v := s.fx.View(t)
	go v.Listen()
	v.FireRuneEvent('Z')
	select {
	case <-v.NextEventProcessed:
		t.Error("unexpected event")
	case <-t.Timeout(1 * time.Millisecond):
	}
	v.Quit()
	t.Eq(0, v.MaxEvents)
}

func (s *View) Reports_quit_event_and_ends_event_loop(t *T) {
	v, quitEvt, terminated := s.fx.View(t), false, false
	v.Listeners.Quit = func() { quitEvt = true }
	go func() {
		v.Listen()
		terminated = true
	}()
	v.FireRuneEvent('q')
	<-v.NextEventProcessed
	t.True(quitEvt)
	t.Eq(-1, v.MaxEvents)
	<-t.Timeout(1 * time.Millisecond) // listener processed before quit
	t.True(terminated)
}

func (s *View) Fails_if_registered_runes_change(t *T) {
	v, err := s.fx.View(t), error(nil)
	v.MaxEvents = 1
	v.Triggers.QuitRunes = []rune{'q', 'Y'}
	v.Listeners.Resize = func(v *lines.View) {
		v.Triggers.QuitRunes = []rune{'q'}
	}
	go func() {
		err = v.Listen()
	}()
	<-v.NextEventProcessed
	v.FireRuneEvent('Y')
	<-v.NextEventProcessed
	t.ErrIs(err, lines.EventRunErr)
}

func TestView(t *testing.T) {
	t.Parallel()
	Run(&View{}, t)
}

type DBG struct{ Suite }

func (s *DBG) Fails_if_registered_runes_change(t *T) {
	v, err := mck.NewView(t), error(nil)
	v.MaxEvents = 1
	v.Triggers.QuitRunes = []rune{'q', 'Y'}
	v.Listeners.Resize = func(v *lines.View) {
		v.Triggers.QuitRunes = []rune{'q'}
	}
	go func() {
		err = v.Listen()
	}()
	<-v.NextEventProcessed
	v.FireRuneEvent('Y')
	<-v.NextEventProcessed
	t.ErrIs(err, lines.EventRunErr)
}

func TestDBG(t *testing.T) { Run(&DBG{}, t) }
