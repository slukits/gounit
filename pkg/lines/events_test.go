package lines

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
)

type NewEvents struct{ Suite }

// setScreenFactory allows to mock up tcell's screen generation for
// error handling testing.  Provided factory instance must implement
// NewScreen() (tcell.Screen, error)
// NewSimulationScreen() tcell.Screen
func setScreenFactory(f screenFactoryer) {
	screenFactory = f
}

func defaultScreenFactory() screenFactoryer {
	return &defaultFactory{}
}

func (s *NewEvents) Fails_if_cell_s_screen_creation_fails(t *T) {
	setScreenFactory(&ScreenFactory{Fail: true})
	_, err := New()
	t.ErrIs(err, ErrScreen)
}

func (s *NewEvents) Fails_if_tcell_s_screen_init_fails(t *T) {
	setScreenFactory(&ScreenFactory{FailInit: true})
	_, err := New()
	t.ErrIs(err, ErrInit)
}

func (s *NewEvents) Succeeds_if_none_of_the_above(t *T) {
	setScreenFactory(&ScreenFactory{})
	_, err := New()
	t.FatalOn(err)
}

func (s *NewEvents) May_fail_in_graphical_test_environment(t *T) {
	// sole purpose of this test is keeping coverage at 100%
	setScreenFactory(defaultScreenFactory())
	ee, err := New()
	if err == nil {
		ee.scr.lib.Fini()
	}
}

func (s *NewEvents) Sim_fails_if_tcell_s_sim_init_fails(t *T) {
	setScreenFactory(&ScreenFactory{FailInit: true})
	_, _, err := Sim()
	t.ErrIs(err, ErrInit)
}

func (s *NewEvents) Sim_succeeds_if_none_of_the_above(t *T) {
	setScreenFactory(defaultScreenFactory())
	_, lib, err := Sim()
	t.FatalOn(err)
	lib.Fini()
}

func (s *NewEvents) Has_copy_of_default_keys_for_internal_events(
	t *T,
) {
	setScreenFactory(defaultScreenFactory())
	ee, _, err := Sim()
	t.FatalOn(err)
	for _, e := range AllFeatures {
		kk := DefaultFeatures.KeysOf(e)
		for _, k := range kk {
			t.True(ee.Features.HasKey(k.Key, k.Mod))
			t.Eq(e, ee.Features.KeyEvent(k.Key, k.Mod))
		}
		rr := DefaultFeatures.RunesOf(e)
		for _, r := range rr {
			t.True(ee.Features.HasRune(r))
			t.Eq(e, ee.Features.RuneEvent(r))
		}
	}
}

func (s *NewEvents) Finalize(t *F) {
	setScreenFactory(defaultScreenFactory())
}

// TestNewRegister can not run in parallel since its tests manipulate the
// package-global state which is necessary to mock errors of the
// tcell-library.
func TestNewEvents(t *testing.T) { Run(&NewEvents{}, t) }

type events struct {
	Suite
	fx *FX
}

func (s *events) Init(t *I) {
	s.fx = NewFX()
}

func (s *events) SetUp(t *T) {
	t.Parallel()
	s.fx.New(t)
}

func (s *events) TearDown(t *T) { s.fx.Del(t) }

func (s *events) Reports_initial_resize_event(t *T) {
	ee, _ := s.fx.For(t)
	resizeListenerCalled := false
	ee.Resize(func(v *Env) { resizeListenerCalled = true })
	ee.Listen()
	t.True(resizeListenerCalled)
}

func (s *events) Stops_reporting_if_view_to_small(t *T) {
	ee, tt := s.fx.For(t, 3)
	updates := 0
	ee.Resize(func(e *Env) { e.SetMin(30) })
	ee.Listen()
	t.Eq(2, tt.Max) // initial resize event
	t.FatalOn(ee.Update(func(v *Env) { updates++ }))
	t.Eq(0, updates)
	tt.FireResize(35)
	t.Eq(1, tt.Max) // second resize event
	t.FatalOn(ee.Update(func(v *Env) { updates++ }))
	t.Eq(1, updates)
}

func (s *events) Stops_reporting_except_for_quit(t *T) {
	ee, tt := s.fx.For(t, -1)
	ee.Resize(func(e *Env) {
		e.SetMin(30)
	})
	tt.FireRune('q')
	t.False(ee.IsListening())
	ee, tt = Test(t.GoT(), -1)
	quit := false
	ee.Resize(func(e *Env) { e.SetMin(30) })
	ee.Quit(func(*Env) { quit = true })
	tt.FireRune('q')
	t.False(ee.IsListening())
	t.True(quit)
}

func (s *events) Posts_and_reports_update_event(t *T) {
	ee, _ := s.fx.For(t)
	update := false
	t.FatalOn(ee.Update(nil))
	t.FatalOn(ee.Update(func(e *Env) { update = true }))
	t.True(update)
	t.False(ee.IsListening())
}

func (s *events) Reports_an_update_event_with_now_timestamp(t *T) {
	ee, _ := s.fx.For(t)
	now, updateReported := time.Now(), false
	ee.Update(func(e *Env) {
		updateReported = true
		t.True(e.Evn.When().After(now))
	})
	t.True(updateReported)
	t.False(ee.IsListening())
}

func (s *events) Fails_posting_an_update_if_event_loop_full(t *T) {
	ee, _, err := Sim()
	t.FatalOn(err)
	block, failed := make(chan struct{}), false
	ee.Update(func(*Env) { <-block })
	for i := 0; i < 100; i++ {
		if err := ee.Update(func(*Env) {}); err != nil {
			failed = true
			break
		}
	}
	close(block)
	ee.QuitListening()
	t.True(failed)
}

func (s *events) Reports_quit_event_and_ends_event_loop(t *T) {
	now := time.Now()
	t.FatalIfNot(t.True(len(DefaultFeatures.KeysOf(FtQuit)) > 0))
	quitKey, quitEvt := DefaultFeatures.KeysOf(FtQuit)[0], false
	ee, tt := s.fx.For(t)
	ee.Quit(func(e *Env) {
		t.True(e.Evn.When().After(now))
		quitEvt = true
	})
	tt.FireKey(quitKey.Key, quitKey.Mod)
	t.True(quitEvt)
	t.Eq(0, tt.Max)
	t.False(ee.IsListening())
}

func (s *events) Quits_event_loop_on_quit_event_without_listener(
	t *T,
) {
	t.FatalIfNot(t.True(len(DefaultFeatures.KeysOf(FtQuit)) > 0))
	quitKey := DefaultFeatures.KeysOf(FtQuit)[0]
	ee, tt := s.fx.For(t)
	ee.Listen()
	tt.FireKey(quitKey.Key, quitKey.Mod)
	t.False(ee.IsListening())
}

func (s *events) Reporting_keyboard_shadows_other_input_listener(
	t *T,
) {
	ee, tt := s.fx.For(t, 2)
	rn, key, kb := false, false, 0
	t.FatalOn(ee.Rune('a', func(*Env) { rn = true }))
	t.FatalOn(ee.Key(tcell.KeyUp, 0, func(*Env) { key = true }))
	ee.Keyboard(func(v *Env, r rune, k tcell.Key, m tcell.ModMask) {
		if r == 'a' {
			kb++
		}
		if k == tcell.KeyUp {
			kb++
		}
	})
	tt.FireRune('a')
	tt.FireKey(tcell.KeyUp)
	t.False(rn)
	t.False(key)
	t.Eq(2, kb)
	t.False(ee.IsListening())
}

func (s *events) Reporting_keyboard_shadows_all_but_quit(t *T) {
	t.FatalIfNot(t.True(len(DefaultFeatures.KeysOf(FtQuit)) > 0))
	quitKey, kb := DefaultFeatures.KeysOf(FtQuit)[0], false
	ee, tt := s.fx.For(t, 2)
	ee.Keyboard(func(e *Env, r rune, k tcell.Key, m tcell.ModMask) {
		kb = true
		ee.Keyboard(nil)
	})
	tt.FireKey(quitKey.Key, quitKey.Mod)
	t.False(ee.IsListening())
	t.False(kb)
}

func (s *events) Stops_reporting_keyboard_if_removed(t *T) {
	ee, tt := s.fx.For(t, 2)
	rn, kb := false, false
	ee.Rune('a', func(*Env) { rn = true })
	ee.Keyboard(func(_ *Env, r rune, k tcell.Key, m tcell.ModMask) {
		t.Eq('a', r)
		kb = true
		ee.Keyboard(nil)
	})
	tt.FireRune('a')
	tt.FireRune('a')
	t.True(kb)
	t.True(rn)
}

func TestEvents(t *testing.T) {
	t.Parallel()
	Run(&events{}, t)
}
