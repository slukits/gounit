// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
)

type _Testing struct{ Suite }

func (s *_Testing) SetUp(t *T) { t.Parallel() }

func (s *_Testing) Starts_non_blocking_listening_with_listen_call(t *T) {
	ee, _ := Test(t.GoT())
	t.False(ee.IsListening())
	ee.Listen()
	t.True(ee.IsListening())
	ee.QuitListening()
}

func (s *_Testing) Starts_listening_if_a_resize_is_fired(t *T) {
	// don't stop after first *reported* event which in this case is the
	// initial resize
	ee, tt := Test(t.GoT(), 2)
	defer ee.QuitListening()
	t.False(ee.IsListening())
	t.True(tt.FireResize(42).IsListening())
}

func (s *_Testing) Starts_listening_if_a_key_is_fired(t *T) {
	ee, tt := Test(t.GoT(), -1) // listen for ever
	defer ee.QuitListening()
	t.False(ee.IsListening())
	t.True(tt.FireKey(tcell.KeyBS, 0).IsListening())
}

func (s *_Testing) Starts_listening_a_rune_is_fired(t *T) {
	ee, tt := Test(t.GoT(), -1)
	defer ee.QuitListening()
	t.False(ee.IsListening())
	t.True(tt.FireRune('r').IsListening())
}

func (s *_Testing) Starts_listening_with_update_call(t *T) {
	ee, _ := Test(t.GoT(), 2) // stop listening after two reported events
	defer ee.QuitListening()
	t.False(ee.IsListening())
	t.FatalOn(ee.Update(func(e *Env) {}))
	t.True(ee.IsListening())
}

func (s *_Testing) Provides_last_screen_after_automatic_quit(t *T) {
	exp := "auto termination"
	ee, tt := Test(t.GoT()) // stops after first reported event
	ee.Resize(func(e *Env) { e.LL().Line(0).Set(exp) })
	ee.Listen() // triggers resize for which we have registered above
	t.False(ee.IsListening())
	t.Eq(exp, fmt.Sprint(tt.LastScreen))
}

func (s *_Testing) Provides_last_screen_after_quit_event(t *T) {
	exp := "event terminated"
	ee, tt := Test(t.GoT(), 3)
	ee.Resize(func(e *Env) { e.LL().Line(0).Set(exp) })
	quitKey := DefaultFeatures.KeysOf(FtQuit)[0]
	tt.FireKey(quitKey.Key, quitKey.Mod)
	t.False(ee.IsListening())
	t.Eq(exp, fmt.Sprint(tt.LastScreen))
}

// TODO: this test fails because of a timeout if go test -count=1000 -race
func (s *_Testing) Provides_last_screen_after_programmatic_quit(t *T) {
	exp := "programmatically terminated"
	ee, tt := Test(t.GoT(), 2)
	// Update will start listening automatically
	ee.Update(func(e *Env) {
		e.LL().Line(0).Set(exp)
		ee.QuitListening()
	})
	t.False(ee.IsListening())
	t.Eq(exp, fmt.Sprint(tt.LastScreen))
}

func (s *_Testing) Fires_initial_resize(t *T) {
	ee, tt := Test(t.GoT())
	ee.Resize(func(e *Env) { e.LL().Line(0).Set("called") })
	ee.Listen()
	t.False(ee.IsListening())
	t.Eq("called", fmt.Sprint(tt.LastScreen))
}

func (s *_Testing) Fires_requested_resize(t *T) {
	ee, tt := Test(t.GoT(), 2)
	resizes, exp := 0, []int{22, 42}
	ee.Resize(func(e *Env) {
		e.LL().Line(resizes).Set(fmt.Sprintf("%d", exp[resizes]))
		resizes++
	})
	tt.FireResize(108)
	t.False(ee.IsListening())
	t.Eq(2, resizes)
	t.Eq("22\n42", fmt.Sprint(tt.LastScreen))
}

func (s *_Testing) Fires_requested_key_event(t *T) {
	ee, tt := Test(t.GoT())
	ee.Key(tcell.KeyEnter, tcell.ModShift, func(e *Env) {
		e.LL().Line(0).Set("shift-enter")
	})
	tt.FireKey(tcell.KeyEnter, tcell.ModShift)
	t.False(ee.IsListening())
	t.Eq("shift-enter", fmt.Sprintf(tt.LastScreen))
}

func (s *_Testing) Fires_requested_rune_event(t *T) {
	ee, tt := Test(t.GoT())
	ee.Rune('a', func(e *Env) { e.LL().Line(0).Set("a-rune") })
	tt.FireRune('a')
	t.False(ee.IsListening())
	t.Eq("a-rune", fmt.Sprintf(tt.LastScreen))
}

func (s *_Testing) Fires_requested_update_event(t *T) {
	ee, tt := Test(t.GoT())
	t.FatalOn(ee.Update(func(e *Env) { e.LL().Line(0).Set("update") }))
	t.False(ee.IsListening())
	t.Eq("update", fmt.Sprintf(tt.LastScreen))
}

func TestTesting(t *testing.T) {
	t.Parallel()
	Run(&_Testing{}, t)
}
