// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
	"github.com/slukits/gounit/pkg/lines/testdata/fx"
)

type NewView struct{ Suite }

func (s *NewView) Fails_if_tcell_s_screen_creation_fails(t *T) {
	lines.SetScreenFactory(&fx.ScreenFactory{Fail: true})
	_, err := lines.NewView()
	t.ErrIs(err, fx.ScreenErr)
}

func (s *NewView) Fails_if_tcell_s_screen_init_fails(t *T) {
	lines.SetScreenFactory(&fx.ScreenFactory{FailInit: true})
	_, err := lines.NewView()
	t.ErrIs(err, fx.InitErr)
}

func (s *NewView) Succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(&fx.ScreenFactory{})
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
	lines.SetScreenFactory(&fx.ScreenFactory{FailInit: true})
	_, _, err := lines.NewSim()
	t.ErrIs(err, fx.InitErr)
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

type AView struct {
	Suite
	fx FX
}

type FX struct {
	*Fixtures
	DefaultLineCount int
}

func (f *FX) View(t *T, maxEvt ...int) *fx.View {
	if len(maxEvt) == 0 {
		return f.Get(t).(*fx.View)
	}
	v := f.Get(t).(*fx.View)
	v.MaxEvents = maxEvt[0]
	return v
}

func (f *FX) Del(t *T) interface{} {
	v := f.Fixtures.Del(t).(*fx.View)
	if v.IsPolling() {
		v.Quit()
	}
	return v
}

func (s *AView) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *AView) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, fx.NewView(t))
}

func (s *AView) Provides_initial_resize_event(t *T) {
	v, resizeListenerCalled := s.fx.View(t), false
	v.Register.Resize(func(v *lines.View) {
		resizeListenerCalled = true
	})
	v.Listen()
	t.True(resizeListenerCalled)
}

func (s *AView) Sim_provides_default_length(t *T) {
	v := s.fx.View(t)
	v.Register.Resize(func(v *lines.View) {
		t.Eq(s.fx.DefaultLineCount, v.Len())
	})
	v.Listen()
}

func (s *AView) Provides_len_many_lines(t *T) {
	v, got := s.fx.View(t), 0
	v.Register.Resize(func(v *lines.View) {
		v.For(func(*lines.Line) { got++ })
	})
	v.Listen()
	t.Eq(s.fx.DefaultLineCount, got)
}

func (s *AView) Provides_nil_line_if_given_index_out_of_bound(t *T) {
	v := s.fx.View(t)
	v.Register.Resize(func(v *lines.View) {
		t.Eq((*lines.Line)(nil), v.Line(v.Len()))
	})
	v.Listen()
}

func (s *AView) Displays_an_error_if_len_to_small(t *T) {
	v, resizeCalled := s.fx.View(t), false
	v.Min = 30
	v.Register.Resize(func(v *lines.View) {
		resizeCalled = true
	})
	go v.Listen()
	<-v.Synced
	t.Contains(v.String(), fmt.Sprintf(lines.ErrScreenFmt, v.Min))
	t.False(resizeCalled)
}

func (s *AView) Resize_adjust_length_accordingly(t *T) {
	v, exp, resizeCount := s.fx.View(t, 1), 20, 0
	v.Register.Resize(func(v *lines.View) {
		switch resizeCount {
		case 0:
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 1:
			t.Eq(exp, v.Len())
		}
		resizeCount++
	})
	go v.Listen()
	<-v.NextEventProcessed // wait for initial resize to happen
	<-v.SetNumberOfLines(exp)
	t.Eq(2, resizeCount)
}

func (s *AView) Increases_lines_count_by_request(t *T) {
	v, exp := s.fx.View(t), 42
	v.Register.Resize(func(v *lines.View) {
		got := 0
		v.ForN(-1, func(l *lines.Line) { got++ })
		t.Eq(0, got)
		v.ForN(exp, func(l *lines.Line) { got++ })
		t.Eq(exp, got)
	})
	go v.Listen()
}

func (s *AView) Resize_adjusts_the_provided_screen_lines(t *T) {
	v, expFirst, expSecond, resizeCount := s.fx.View(t, 3), 15, 30, 0
	v.Register.Resize(func(v *lines.View) {
		switch resizeCount {
		case 0:
			got := 0
			v.ForScreen(func(*lines.Line) { got++ })
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 1:
			got := 0
			v.ForScreen(func(*lines.Line) { got++ })
			t.Eq(s.fx.DefaultLineCount, v.Len())
		case 2:
			got := 0
			v.ForScreen(func(*lines.Line) { got++ })
			t.Eq(expFirst, got)
		case 3:
			got := 0
			v.ForScreen(func(*lines.Line) { got++ })
			t.Eq(expSecond, got)
		}
		resizeCount++
	})
	go v.Listen()
	<-v.NextEventProcessed
	<-v.SetNumberOfLines(s.fx.DefaultLineCount)
	<-v.SetNumberOfLines(expFirst)
	<-v.SetNumberOfLines(expSecond)
	t.Eq(4, resizeCount)
}

func (s *AView) Reports_quit_event_and_ends_event_loop(t *T) {
	quit := []int{int('q'), int(tcell.KeyCtrlC), int(tcell.KeyCtrlD)}
	for i, k := range quit {
		v, quitEvt, terminated := fx.NewView(t, 1), false, false
		v.Register.Quit(func() { quitEvt = true })
		go func() {
			v.Listen()
			terminated = true
		}()
		if i == 0 {
			v.FireRuneEvent(rune(k))
		} else {
			v.FireKeyEvent(tcell.Key(k))
		}
		<-v.NextEventProcessed
		t.True(quitEvt)
		<-t.Timeout(1 * time.Millisecond) // listener processed before quit
		t.True(terminated)
		t.Eq(0, v.MaxEvents)
	}
}

func (s *AView) Quits_event_loop_on_quit_event_without_listener(t *T) {
	v, terminated := s.fx.View(t), true
	go func() {
		v.Listen()
		terminated = true
	}()
	v.FireRuneEvent('q') // here we can not wait on the event!!
	<-t.Timeout(1 * time.Millisecond)
	t.True(terminated)
}

func (s *AView) Unregistered_events_are_ignored(t *T) {
	v := s.fx.View(t)
	go v.Listen()
	v.FireRuneEvent('a')
	select {
	case <-v.NextEventProcessed:
		t.Error("unexpected event")
	case <-t.Timeout(1 * time.Millisecond):
	}
	v.FireKeyEvent(tcell.KeyF11)
	select {
	case <-v.NextEventProcessed:
		t.Error("unexpected event")
	case <-t.Timeout(1 * time.Millisecond):
	}
	v.Quit()
	t.Eq(0, v.MaxEvents)
}

func (s *AView) Reports_registered_rune_and_key_events(t *T) {
	v, shiftEnter, aRune := s.fx.View(t, 1), false, false
	err := v.Register.Key(func(v *lines.View, m tcell.ModMask) {
		if m == tcell.ModShift {
			shiftEnter = true
		}
	}, tcell.KeyEnter)
	go v.Listen()
	t.FatalOn(err)
	t.FatalOn(v.Register.Rune(func(v *lines.View) { aRune = true }, 'a'))
	<-v.FireKeyEvent(tcell.KeyEnter, tcell.ModShift)
	t.True(shiftEnter)
	<-v.FireRuneEvent('a')
	t.True(aRune)
	t.Eq(-1, v.MaxEvents)
}

func (s *AView) Unregisters_nil_listener_events(t *T) {
	v := s.fx.View(t)
	defer v.Quit()
	t.FatalOn(v.Register.Rune(func(*lines.View) {}, 'a'))
	t.FatalOn(v.Register.Rune(nil, 'a'))
	t.FatalOn(v.Register.Rune(func(*lines.View) {}, 'a'))
	t.FatalOn(v.Register.Key(
		func(*lines.View, tcell.ModMask) {}, tcell.KeyUp))
	t.FatalOn(v.Register.Key(nil, tcell.KeyUp))
	t.FatalOn(v.Register.Key(
		func(*lines.View, tcell.ModMask) {}, tcell.KeyUp))
}

func (s *AView) Fails_to_register_overwriting_key_or_rune_events(t *T) {
	v, fail := s.fx.View(t), []int{int('a'), int('q'), int(tcell.KeyUp),
		int(tcell.KeyCtrlC), int(tcell.KeyCtrlD)}
	defer v.Quit()
	t.FatalOn(v.Register.Rune(func(*lines.View) {}, 'a'))
	err := v.Register.Key(
		func(*lines.View, tcell.ModMask) {}, tcell.KeyUp)
	t.FatalOn(err)
	for i, k := range fail {
		switch i {
		case 0, 1:
			t.ErrIs(
				v.Register.Rune(func(*lines.View) {}, rune(k)),
				lines.RegisterErr,
			)
		default:
			t.ErrIs(v.Register.Key(
				func(*lines.View, tcell.ModMask) {}, tcell.Key(k)),
				lines.RegisterErr,
			)
		}
	}
}

func (s *AView) Reports_all_rune_events_to_runes_listener_til_removed(
	t *T,
) {
	v, aRune, allRunes := s.fx.View(t, 1), false, false
	v.Register.Rune(func(v *lines.View) { aRune = true }, 'a')
	v.Register.Runes(func(v *lines.View, r rune) { allRunes = true })
	go v.Listen()
	<-v.FireRuneEvent('a')
	t.True(allRunes)
	v.Register.Runes(nil)
	<-v.FireRuneEvent('a')
	t.True(aRune)
	t.Eq(-1, v.MaxEvents)
}

func TestAView(t *testing.T) {
	t.Parallel()
	Run(&AView{}, t)
}
