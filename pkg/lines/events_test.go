package lines_test

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

type NewEvents struct{ Suite }

func (s *NewEvents) Fails_if_cell_s_screen_creation_fails(t *T) {
	lines.SetScreenFactory(&ScreenFactory{Fail: true})
	_, err := lines.New()
	t.ErrIs(err, ErrScreen)
}

func (s *NewEvents) Fails_if_tcell_s_screen_init_fails(t *T) {
	lines.SetScreenFactory(&ScreenFactory{FailInit: true})
	_, err := lines.New()
	t.ErrIs(err, ErrInit)
}

func (s *NewEvents) Succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(&ScreenFactory{})
	_, err := lines.New()
	t.FatalOn(err)
}

func (s *NewEvents) May_fail_in_graphical_test_environment(t *T) {
	// sole purpose of this test is keeping coverage at 100%
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	rg, err := lines.New()
	if err == nil {
		lines.GetLib(rg).Fini()
	}
}

func (s *NewEvents) Sim_fails_if_tcell_s_sim_init_fails(t *T) {
	lines.SetScreenFactory(&ScreenFactory{FailInit: true})
	_, _, err := lines.Sim()
	t.ErrIs(err, ErrInit)
}

func (s *NewEvents) Sim_succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	_, lib, err := lines.Sim()
	t.FatalOn(err)
	lib.Fini()
}

func (s *NewEvents) Has_copy_of_default_keys_for_internal_events(
	t *T,
) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	reg, _, err := lines.Sim()
	t.FatalOn(err)
	for _, e := range lines.AllFeatures {
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

func (s *NewEvents) Finalize(t *F) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
}

// TestNewRegister can not run in parallel since its tests manipulate the
// package-global state which is necessary to mock errors of the
// tcell-library.
func TestNewRegister(t *testing.T) { Run(&NewEvents{}, t) }

type events struct {
	Suite
	fx FX
}

type FX struct {
	*Fixtures
	DefaultLineCount int
}

func (f *FX) EE(t *T, maxEvt ...int) *Events {
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

func (s *events) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *events) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, New(t))
}

func (s *events) TearDown(t *T) { s.fx.Del(t) }

func (s *events) Reports_initial_resize_event(t *T) {
	ee, resizeListenerCalled := s.fx.EE(t), false
	ee.Resize(func(v *lines.Screen) { resizeListenerCalled = true })
	ee.Listen()
	t.True(resizeListenerCalled)
}

func (s *events) Stops_reporting_if_view_to_small(t *T) {
	ee, updates := s.fx.EE(t, 2), 0
	ee.Resize(func(v *lines.Screen) {
		v.SetMin(30)
	})
	ee.Listen()
	t.Eq(1, ee.Max) // initial resize event
	t.FatalOn(ee.Update(func(v *lines.Screen) { updates++ }))
	t.Eq(0, updates)
	ee.SetNumberOfLines(35)
	t.Eq(0, ee.Max) // second resize event
	t.FatalOn(ee.Update(func(v *lines.Screen) { updates++ }))
	t.Eq(1, updates)
}

func (s *events) Stops_reporting_except_for_quit(t *T) {
	ee := s.fx.EE(t, 2)
	ee.Resize(func(v *lines.Screen) {
		v.SetMin(30)
	})
	ee.Listen()
	ee.FireRuneEvent('q')
	t.False(ee.IsPolling())
	ee, quit := New(t, 2), false
	ee.Resize(func(v *lines.Screen) {
		v.SetMin(30)
	})
	ee.Quit(func() { quit = true })
	ee.Listen()
	ee.FireRuneEvent('q')
	t.False(ee.IsPolling())
	t.True(quit)
}

func (s *events) Posts_and_reports_update_event(t *T) {
	ee, update := s.fx.EE(t), false
	t.FatalOn(ee.Update(nil))
	t.FatalOn(ee.Update(func(v *lines.Screen) {
		update = true
	}))
	t.True(update)
	t.False(ee.IsPolling())
}

func (s *events) Reports_an_update_event_with_now_timestamp(t *T) {
	ee, now, updateReported := s.fx.EE(t, 0), time.Now(), false
	ee.Update(func(v *lines.Screen) {
		updateReported = true
		t.True(ee.Ev.When().After(now))
	})
	t.True(updateReported)
	t.False(ee.IsPolling())
}
func (s *events) Fails_posting_an_update_if_event_loop_full(t *T) {
	ee, _, err := lines.Sim()
	t.FatalOn(err)
	block, failed := make(chan struct{}), false
	ee.Update(func(v *lines.Screen) { <-block })
	for i := 0; i < 100; i++ {
		if err := ee.Update(func(v *lines.Screen) {}); err != nil {
			failed = true
			break
		}
	}
	close(block)
	ee.QuitListening()
	t.True(failed)
}

func (s *events) Reports_quit_event_and_ends_event_loop(t *T) {
	quit := []int{0, int('q'), int(tcell.KeyCtrlC), int(tcell.KeyCtrlD)}
	now := time.Now()
	for i, k := range quit {
		ee, quitEvt := New(t, 1), false
		ee.Quit(func() {
			if i == 0 {
				t.True(ee.Ev.When().After(now))
			}
			quitEvt = true
		})
		ee.Listen()
		switch i {
		case 0:
			ee.QuitListening()
		case 1:
			ee.FireRuneEvent(rune(k))
		default:
			ee.FireKeyEvent(tcell.Key(k))
		}
		t.True(quitEvt)
		if i == 0 {
			t.Eq(1, ee.Max)
		} else {
			t.Eq(0, ee.Max)
		}
		t.False(ee.IsPolling())
	}
}

func (s *events) Quits_event_loop_on_quit_event_without_listener(
	t *T,
) {
	ee := s.fx.EE(t)
	ee.Listen()
	ee.FireRuneEvent('q')
	t.False(ee.IsPolling())
}

func (s *events) Reporting_keyboard_shadows_other_input_listener(
	t *T,
) {
	ee, rn, key, kb := s.fx.EE(t, 1), false, false, 0
	t.FatalOn(ee.Rune('a', func(v *lines.Screen) { rn = true }))
	t.FatalOn(ee.Key(tcell.KeyUp, 0, func(v *lines.Screen) {
		key = true
	}))
	ee.Keyboard(func(
		v *lines.Screen, r rune, k tcell.Key, m tcell.ModMask,
	) {
		if r == 'a' {
			kb++
		}
		if k == tcell.KeyUp {
			kb++
		}
	})
	ee.FireRuneEvent('a')
	ee.FireKeyEvent(tcell.KeyUp)
	t.False(rn)
	t.False(key)
	t.Eq(2, kb)
	t.False(ee.IsPolling())
}

func (s *events) Reporting_keyboard_shadows_all_but_quit(t *T) {
	ee, kb := s.fx.EE(t, 1), false
	ee.Keyboard(func(
		v *lines.Screen, r rune, k tcell.Key, m tcell.ModMask,
	) {
		kb = true
		ee.Keyboard(nil)
	})
	ee.FireKeyEvent(tcell.KeyCtrlC)
	t.False(ee.IsPolling())
	t.False(kb)
}

func (s *events) Stops_reporting_keyboard_if_removed(t *T) {
	ee, rn, kb := s.fx.EE(t, 1), false, false
	ee.Rune('a', func(v *lines.Screen) { rn = true })
	ee.Keyboard(func(
		v *lines.Screen, r rune, k tcell.Key, m tcell.ModMask,
	) {
		t.Eq('a', r)
		kb = true
		ee.Keyboard(nil)
	})
	ee.FireRuneEvent('a')
	ee.FireRuneEvent('a')
	t.True(kb)
	t.True(rn)
	t.False(ee.IsPolling())
}

func TestARegister(t *testing.T) {
	t.Parallel()
	Run(&events{}, t)
}
