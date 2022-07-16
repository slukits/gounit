package lines_test

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

type NewRegister struct{ Suite }

func (s *NewRegister) Fails_if_cell_s_screen_creation_fails(t *T) {
	lines.SetScreenFactory(&ScreenFactory{Fail: true})
	_, err := lines.New()
	t.ErrIs(err, ErrScreen)
}

func (s *NewRegister) Fails_if_tcell_s_screen_init_fails(t *T) {
	lines.SetScreenFactory(&ScreenFactory{FailInit: true})
	_, err := lines.New()
	t.ErrIs(err, ErrInit)
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
	t.ErrIs(err, ErrInit)
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
	for _, e := range lines.InternalFeatures {
		kk := lines.DefaultFeatures.KeysOf(e)
		for _, k := range kk {
			t.True(reg.Features.HasKey(k.Key, k.Mod))
			t.Eq(e, reg.Features.KeyEvent(k.Key, k.Mod))
		}
		rr := lines.DefaultFeatures.RunesOf(e)
		for _, r := range rr {
			t.True(reg.Features.HasRune(r))
			t.Eq(e, reg.Features.RuneEvent(r))
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

func (f *FX) Reg(t *T, maxEvt ...int) *Events {
	if len(maxEvt) == 0 {
		return f.Get(t).(*Events)
	}
	rg := f.Get(t).(*Events)
	rg.Max = maxEvt[0]
	return rg
}

func (f *FX) Del(t *T) interface{} {
	rg, ok := f.Fixtures.Del(t).(*Events)
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
	rg := s.fx.Reg(t, 2)
	rg.Timeout = 5 * time.Minute
	rg.Resize(func(v *lines.View) {
		v.SetMin(30)
	})
	rg.Listen()
	rg.FireRuneEvent('q')
	t.False(rg.IsPolling())
	rg, quit := New(t, 2), false
	rg.Resize(func(v *lines.View) {
		v.SetMin(30)
	})
	rg.Quit(func() { quit = true })
	rg.Listen()
	rg.FireRuneEvent('q')
	t.False(rg.IsPolling())
	t.True(quit)
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
		switch i {
		case 0:
			rg.QuitListening()
		case 1:
			rg.FireRuneEvent(rune(k))
		default:
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

func (s *ARegister) Reporting_keyboard_shadows_other_input_listener(
	t *T,
) {
	rg, rn, key, kb := s.fx.Reg(t, 1), false, false, 0
	t.FatalOn(rg.Rune('a', func(v *lines.View) { rn = true }))
	t.FatalOn(rg.Key(tcell.KeyUp, 0, func(v *lines.View) {
		key = true
	}))
	rg.Keyboard(func(
		v *lines.View, r rune, k tcell.Key, m tcell.ModMask,
	) {
		if r == 'a' {
			kb++
		}
		if k == tcell.KeyUp {
			kb++
		}
	})
	rg.Listen()
	rg.FireRuneEvent('a')
	rg.FireKeyEvent(tcell.KeyUp)
	t.False(rn)
	t.False(key)
	t.Eq(2, kb)
	t.False(rg.IsPolling())
}

func (s *ARegister) Reporting_keyboard_shadows_all_but_quit(t *T) {
	rg, kb := s.fx.Reg(t, 1), false
	rg.Keyboard(func(
		v *lines.View, r rune, k tcell.Key, m tcell.ModMask,
	) {
		kb = true
		rg.Keyboard(nil)
	})
	rg.Listen()
	rg.FireKeyEvent(tcell.KeyCtrlC)
	t.False(rg.IsPolling())
	t.False(kb)
}

func (s *ARegister) Stops_reporting_keyboard_if_removed(t *T) {
	rg, rn, kb := s.fx.Reg(t, 1), false, false
	rg.Rune('a', func(v *lines.View) { rn = true })
	rg.Keyboard(func(
		v *lines.View, r rune, k tcell.Key, m tcell.ModMask,
	) {
		t.Eq('a', r)
		kb = true
		rg.Keyboard(nil)
	})
	rg.Listen()
	rg.FireRuneEvent('a')
	rg.FireRuneEvent('a')
	t.True(kb)
	t.True(rn)
	t.False(rg.IsPolling())
}

func TestARegister(t *testing.T) {
	t.Parallel()
	Run(&ARegister{}, t)
}
