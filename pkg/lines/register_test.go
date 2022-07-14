package lines_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

func waitFor(t *T, c interface{}, mm ...string) {
	msg := func() string {
		if len(mm) == 0 {
			return "timed out"
		}
		return fmt.Sprintf("timed out: %s", mm[0])
	}
	t.GoT().Helper()
	switch c := c.(type) {
	case chan struct{}:
		select {
		case <-c:
		case <-t.Timeout(10 * time.Millisecond):
			t.Fatalf(msg())
		}
	case chan bool:
		select {
		case <-c:
		case <-t.Timeout(10 * time.Millisecond):
			t.Fatalf(msg())
		}
	}
}

type NewRegister struct{ Suite }

func (s *NewRegister) Fails_if_cell_s_screen_creation_fails(t *T) {
	lines.SetScreenFactory(&ScreenFactory{Fail: true})
	_, err := lines.New()
	t.ErrIs(err, ScreenErr)
}

func (s *NewRegister) Fails_if_tcell_s_screen_init_fails(t *T) {
	lines.SetScreenFactory(&ScreenFactory{FailInit: true})
	_, err := lines.New()
	t.ErrIs(err, InitErr)
}

func (s *NewRegister) Succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(&ScreenFactory{})
	_, err := lines.New()
	t.FatalOn(err)
}

func (s *NewRegister) May_fail_in_graphical_test_environment(t *T) {
	// sole purpose of this test is keeping coverage at 100%
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	rg, err := lines.New()
	if err == nil {
		lines.GetLib(rg).Fini()
	}
}

func (s *NewRegister) Sim_fails_if_tcell_s_sim_init_fails(t *T) {
	lines.SetScreenFactory(&ScreenFactory{FailInit: true})
	_, _, err := lines.Sim()
	t.ErrIs(err, InitErr)
}

func (s *NewRegister) Sim_succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	_, lib, err := lines.Sim()
	t.FatalOn(err)
	lib.Fini()
}

func (s *NewRegister) Has_copy_of_default_keys_for_internal_events(
	t *T,
) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	reg, _, err := lines.Sim()
	t.FatalOn(err)
	for _, e := range lines.InternalEvents {
		kk := lines.DefaultKeys.KeysOf(e)
		for _, k := range kk {
			t.True(reg.Keys.HasKey(k))
			t.Eq(e, reg.Keys.KeyEvent(k))
		}
		rr := lines.DefaultKeys.RunesOf(e)
		for _, r := range rr {
			t.True(reg.Keys.HasRune(r))
			t.Eq(e, reg.Keys.RuneEvent(r))
		}
	}
}

func (s *NewRegister) Finalize(t *F) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
}

// TestNewRegister can not run in parallel since its tests manipulate the
// package-global state which is necessary to mock errors of the
// tcell-library.
func TestNewRegister(t *testing.T) { Run(&NewRegister{}, t) }

type ARegister struct {
	Suite
	fx FX
}

type FX struct {
	*Fixtures
	DefaultLineCount int
}

func (f *FX) Reg(t *T, maxEvt ...int) *Register {
	if len(maxEvt) == 0 {
		return f.Get(t).(*Register)
	}
	rg := f.Get(t).(*Register)
	rg.Max = maxEvt[0]
	return rg
}

func (f *FX) Del(t *T) interface{} {
	rg, ok := f.Fixtures.Del(t).(*Register)
	if !ok {
		return nil
	}
	if rg.IsPolling() {
		rg.QuitListening()
	}
	return rg
}

func (s *ARegister) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *ARegister) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, New(t))
}

func (s *ARegister) TearDown(t *T) { s.fx.Del(t) }

func (s *ARegister) Reports_initial_resize_event(t *T) {
	rg, resizeListenerCalled := s.fx.Reg(t), false
	rg.Resize(func(v *lines.View) { resizeListenerCalled = true })
	rg.Listen()
	t.True(resizeListenerCalled)
}

func (s *ARegister) Stops_reporting_if_view_to_small(t *T) {
	rg, updates := s.fx.Reg(t, 2), 0
	rg.Resize(func(v *lines.View) {
		v.SetMin(30)
	})
	rg.Listen()
	t.Eq(1, rg.Max) // initial resize event
	t.FatalOn(rg.Update(func(v *lines.View) { updates++ }))
	t.Eq(0, updates)
	rg.SetNumberOfLines(35)
	t.Eq(0, rg.Max) // second resize event
	t.FatalOn(rg.Update(func(v *lines.View) { updates++ }))
	t.Eq(1, updates)
}

func (s *ARegister) Stops_reporting_except_for_quit(t *T) {
	rg := s.fx.Reg(t, 1)
	rg.Resize(func(v *lines.View) {
		v.SetMin(30)
	})
	rg.Listen()
	rg.FireRuneEvent('q')
	t.False(rg.IsPolling())
}

func (s *ARegister) Posts_and_reports_update_event(t *T) {
	rg, update := s.fx.Reg(t), false
	rg.Listen()
	t.FatalOn(rg.Update(nil))
	t.FatalOn(rg.Update(func(v *lines.View) {
		update = true
	}))
	t.True(update)
	t.False(rg.IsPolling())
}

func (s *ARegister) Reports_an_update_event_with_now_timestamp(t *T) {
	rg, now, updateReported := s.fx.Reg(t, 0), time.Now(), false
	rg.Listen()
	rg.Update(func(v *lines.View) {
		updateReported = true
		t.True(rg.Ev.When().After(now))
	})
	t.True(updateReported)
	t.False(rg.IsPolling())
}
func (s *ARegister) Fails_posting_an_update_if_event_loop_full(t *T) {
	rg, _, err := lines.Sim()
	t.FatalOn(err)
	block, failed := make(chan struct{}), false
	rg.Update(func(v *lines.View) { <-block })
	for i := 0; i < 100; i++ {
		if err := rg.Update(func(v *lines.View) {}); err != nil {
			failed = true
			break
		}
	}
	close(block)
	rg.QuitListening()
	t.True(failed)
}

func (s *ARegister) Reports_quit_event_and_ends_event_loop(t *T) {
	quit := []int{0, int('q'), int(tcell.KeyCtrlC), int(tcell.KeyCtrlD)}
	now := time.Now()
	for i, k := range quit {
		rg, quitEvt := New(t, 1), false
		rg.Quit(func() {
			if i == 0 {
				t.True(rg.Ev.When().After(now))
			}
			quitEvt = true
		})
		rg.Listen()
		if i == 0 {
			rg.QuitListening()
		} else if i == 1 {
			rg.FireRuneEvent(rune(k))
		} else {
			rg.FireKeyEvent(tcell.Key(k))
		}
		t.True(quitEvt)
		t.Eq(0, rg.Max)
		t.False(rg.IsPolling())
	}
}

func (s *ARegister) Quits_event_loop_on_quit_event_without_listener(
	t *T,
) {
	rg := s.fx.Reg(t)
	rg.Listen()
	rg.FireRuneEvent('q')
	t.False(rg.IsPolling())
}

func (s *ARegister) Does_not_report_unregistered_events(t *T) {
	rg := s.fx.Reg(t)
	rg.Listen()
	rg.FireRuneEvent('a')
	rg.FireKeyEvent(tcell.KeyF11)
	t.Eq(0, rg.Max)
}

func (s *ARegister) Reports_registered_rune_and_key_events(t *T) {
	rg, shiftEnter, aRune := s.fx.Reg(t, 1), false, false
	err := rg.Key(func(v *lines.View, m tcell.ModMask) {
		if m == tcell.ModShift {
			shiftEnter = true
		}
	}, tcell.KeyEnter)
	t.FatalOn(err)
	t.FatalOn(rg.Rune(func(v *lines.View) { aRune = true }, 'a'))
	rg.Listen()
	rg.FireKeyEvent(tcell.KeyEnter, tcell.ModShift)
	t.True(shiftEnter)
	rg.FireRuneEvent('a')
	t.True(aRune)
	t.Eq(-1, rg.Max)
}

func (s *ARegister) Unregisters_nil_listener_events(t *T) {
	rg := s.fx.Reg(t)
	t.FatalOn(rg.Rune(func(*lines.View) {}, 'a'))
	t.FatalOn(rg.Rune(nil, 'a'))
	t.FatalOn(rg.Rune(func(*lines.View) {}, 'a'))
	t.FatalOn(rg.Key(
		func(*lines.View, tcell.ModMask) {}, tcell.KeyUp))
	t.FatalOn(rg.Key(nil, tcell.KeyUp))
	t.FatalOn(rg.Key(
		func(*lines.View, tcell.ModMask) {}, tcell.KeyUp))
}

func (s *ARegister) Fails_to_register_overwriting_key_or_rune_events(
	t *T,
) {
	rg, fail := s.fx.Reg(t), []int{int('a'), int('q'), int(tcell.KeyUp),
		int(tcell.KeyCtrlC), int(tcell.KeyCtrlD)}
	defer rg.QuitListening()
	t.FatalOn(rg.Rune(func(*lines.View) {}, 'a'))
	err := rg.Key(func(*lines.View, tcell.ModMask) {}, tcell.KeyUp)
	t.FatalOn(err)
	for i, k := range fail {
		switch i {
		case 0, 1:
			t.ErrIs(
				rg.Rune(func(*lines.View) {}, rune(k)),
				lines.RegisterErr,
			)
		default:
			t.ErrIs(rg.Key(
				func(*lines.View, tcell.ModMask) {}, tcell.Key(k)),
				lines.RegisterErr,
			)
		}
	}
}

func (s *ARegister) Reports_all_rune_events_to_runes_listener_til_removed(
	t *T,
) {
	rg, aRune, allRunes := s.fx.Reg(t, 1), false, false
	rg.Rune(func(v *lines.View) { aRune = true }, 'a')
	rg.Runes(func(v *lines.View, r rune) { allRunes = true })
	rg.Listen()
	rg.FireRuneEvent('a')
	t.True(allRunes)
	rg.Runes(nil)
	rg.FireRuneEvent('a')
	t.True(aRune)
	t.False(rg.IsPolling())
}

func TestARegister(t *testing.T) {
	t.Parallel()
	Run(&ARegister{}, t)
}
