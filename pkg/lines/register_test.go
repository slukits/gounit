package lines_test

import (
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
	"github.com/slukits/gounit/pkg/lines/testdata/fx"
)

type NewRegister struct{ Suite }

func (s *NewRegister) Fails_if_cell_s_screen_creation_fails(t *T) {
	lines.SetScreenFactory(&fx.ScreenFactory{Fail: true})
	_, err := lines.New()
	t.ErrIs(err, fx.ScreenErr)
}

func (s *NewRegister) Fails_if_tcell_s_screen_init_fails(t *T) {
	lines.SetScreenFactory(&fx.ScreenFactory{FailInit: true})
	_, err := lines.New()
	t.ErrIs(err, fx.InitErr)
}

func (s *NewRegister) Succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(&fx.ScreenFactory{})
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
	lines.SetScreenFactory(&fx.ScreenFactory{FailInit: true})
	_, _, err := lines.Sim()
	t.ErrIs(err, fx.InitErr)
}

func (s *NewRegister) Sim_succeeds_if_none_of_the_above(t *T) {
	lines.SetScreenFactory(lines.DefaultScreenFactory())
	_, lib, err := lines.Sim()
	t.FatalOn(err)
	lib.Fini()
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

func (f *FX) Reg(t *T, maxEvt ...int) *fx.Register {
	if len(maxEvt) == 0 {
		return f.Get(t).(*fx.Register)
	}
	rg := f.Get(t).(*fx.Register)
	rg.Max = maxEvt[0]
	return rg
}

func (f *FX) Del(t *T) interface{} {
	rg, ok := f.Fixtures.Del(t).(*fx.Register)
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
	s.fx.Set(t, fx.New(t))
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
	go rg.Listen()
	<-rg.NextEventProcessed
	t.FatalOn(rg.Update(func(v *lines.View) { updates++ }))
	<-rg.Synced
	<-rg.SetNumberOfLines(35)
	t.FatalOn(rg.Update(func(v *lines.View) { updates++ }))
	<-rg.NextEventProcessed
}

func (s *ARegister) Stops_reporting_except_for_quit(t *T) {
	rg := s.fx.Reg(t, 1)
	rg.Resize(func(v *lines.View) {
		v.SetMin(30)
	})
	go rg.Listen()
	<-rg.NextEventProcessed
	rg.FireRuneEvent('q')
	<-rg.NextEventProcessed
	t.False(rg.IsPolling())
}

func (s *ARegister) Posts_and_reports_update_event(t *T) {
	rg, update := s.fx.Reg(t, 1), false
	t.FatalOn(rg.Update(nil))
	t.FatalOn(rg.Update(func(v *lines.View) { update = true }))
	go rg.Listen()
	<-rg.NextEventProcessed
	rg.QuitListening()
	t.True(update)
	t.Eq(0, rg.Max) // i.e. only one event was reported
}

func (s *ARegister) Reports_a_update_event_with_now_timestamp(t *T) {
	rg, now, updateReported := s.fx.Reg(t, 0), time.Now(), false
	go rg.Listen()
	rg.Update(func(v *lines.View) {
		updateReported = true
		t.True(rg.Ev.When().After(now))
	})
	<-rg.NextEventProcessed
	t.True(updateReported)
}

func (s *ARegister) Fails_posting_an_update_if_event_loop_full(t *T) {
	rg, _, err := lines.Sim()
	t.FatalOn(err)
	block, failed := make(chan struct{}), false
	// go rg.Listen()
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
	quit := []int{int('q'), int(tcell.KeyCtrlC), int(tcell.KeyCtrlD)}
	for i, k := range quit {
		rg, quitEvt, terminated := fx.New(t, 1), false, false
		rg.Quit(func() { quitEvt = true })
		go func() {
			rg.Listen()
			terminated = true
		}()
		if i == 0 {
			rg.FireRuneEvent(rune(k))
		} else {
			rg.FireKeyEvent(tcell.Key(k))
		}
		<-rg.NextEventProcessed
		t.True(quitEvt)
		t.True(terminated)
		t.Eq(0, rg.Max)
	}
}

func (s *ARegister) Quits_event_loop_on_quit_event_without_listener(
	t *T,
) {
	rg, terminated := s.fx.Reg(t), make(chan struct{})
	go func() {
		rg.Listen()
		close(terminated)
	}()
	rg.FireRuneEvent('q') // here we can not wait on the event!!
	select {
	case <-t.Timeout(10 * time.Millisecond):
		t.Error("register listener doesn't seem to terminate")
	case <-terminated:
	}
}

func (s *ARegister) Does_not_report_unregistered_events(t *T) {
	rg := s.fx.Reg(t)
	go rg.Listen()
	rg.FireRuneEvent('a')
	select {
	case <-rg.NextEventProcessed:
		t.Error("unexpected event")
	case <-t.Timeout(1 * time.Millisecond):
	}
	rg.FireKeyEvent(tcell.KeyF11)
	select {
	case <-rg.NextEventProcessed:
		t.Error("unexpected event")
	case <-t.Timeout(1 * time.Millisecond):
	}
	t.Eq(0, rg.Max)
}

func (s *ARegister) Reports_registered_rune_and_key_events(t *T) {
	rg, shiftEnter, aRune := s.fx.Reg(t, 1), false, false
	err := rg.Key(func(v *lines.View, m tcell.ModMask) {
		if m == tcell.ModShift {
			shiftEnter = true
		}
	}, tcell.KeyEnter)
	go rg.Listen()
	t.FatalOn(err)
	t.FatalOn(rg.Rune(func(v *lines.View) { aRune = true }, 'a'))
	<-rg.FireKeyEvent(tcell.KeyEnter, tcell.ModShift)
	t.True(shiftEnter)
	<-rg.FireRuneEvent('a')
	t.True(aRune)
	t.Eq(-1, rg.Max)
}

func (s *ARegister) Unregisters_nil_listener_events(t *T) {
	rg := s.fx.Reg(t)
	defer rg.QuitListening()
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
	go rg.Listen()
	<-rg.FireRuneEvent('a')
	t.True(allRunes)
	rg.Runes(nil)
	<-rg.FireRuneEvent('a')
	t.True(aRune)
	t.Eq(-1, rg.Max)
}

func TestARegister(t *testing.T) {
	t.Parallel()
	Run(&ARegister{}, t)
}
