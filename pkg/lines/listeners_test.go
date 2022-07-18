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

type listeners struct {
	Suite
	fx *LLFixtures
}

type LLFixtures struct{ Fixtures }

func (ll *LLFixtures) LL(t *T) *Listeners {
	return ll.Get(t).(*Listeners)
}

func (s *listeners) Init(t *I) { s.fx = &LLFixtures{} }

func (s *listeners) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, NewListeners(DefaultFeatures))
}

func (s *listeners) Has_initially_no_keyboard_listener(t *T) {
	t.False(s.fx.LL(t).HasKBListener())
}

func (s *listeners) Is_not_ok_if_listener_for_missing_key_requested(t *T) {
	_, ok := s.fx.LL(t).KeyListenerOf(tcell.KeyEnter, 0)
	t.False(ok)
}

func (s *listeners) Is_not_ok_if_listener_for_missing_rune_requested(t *T) {
	_, ok := s.fx.LL(t).RuneListenerOf('r')
	t.False(ok)
}

var zeroListener = func(*Env) {}

func (s *listeners) Fails_to_add_the_zero_rune(t *T) {
	t.ErrIs(s.fx.LL(t).Rune(0, zeroListener), ErrZeroRune)
}

func (s *listeners) Fails_to_add_rune_associated_with_quit_feature(t *T) {
	err := s.fx.LL(t).Rune(
		DefaultFeatures.RunesOf(FtQuit)[0], zeroListener)
	t.ErrIs(err, ErrQuit)
}

func (s *listeners) Fails_to_add_already_registered_rune(t *T) {
	kk := s.fx.LL(t)
	t.FatalOn(kk.Rune('a', zeroListener))
	t.ErrIs(kk.Rune('a', zeroListener), ErrExists)
}

func (s *listeners) Adds_rune_event_if_none_of_the_above(t *T) {
	kk := s.fx.LL(t)
	t.FatalOn(kk.Rune('a', zeroListener))
	l, ok := kk.RuneListenerOf('a')
	t.True(ok)
	t.Eq(fmt.Sprintf("%p", zeroListener), fmt.Sprintf("%p", l))
}

func (s *listeners) Removes_rune_event_if_given_listener_nil(t *T) {
	kk := s.fx.LL(t)
	t.FatalOn(kk.Rune('a', zeroListener))
	t.FatalOn(kk.Rune('a', nil))
	t.FatalOn(kk.Rune('a', zeroListener))
}

func (s *listeners) Fails_to_add_the_zero_key(t *T) {
	t.ErrIs(s.fx.LL(t).Key(0, 0, zeroListener), ErrZeroKey)
}

func (s *listeners) Fails_to_add_key_associated_with_quit_feature(t *T) {
	err := s.fx.LL(t).Key(
		DefaultFeatures.KeysOf(FtQuit)[0].Key, 0, zeroListener)
	t.ErrIs(err, ErrQuit)
}

func (s *listeners) Fails_to_add_already_registered_key(t *T) {
	kk := s.fx.LL(t)
	t.FatalOn(kk.Key(tcell.KeyBS, 0, zeroListener))
	t.ErrIs(kk.Key(tcell.KeyBS, 0, zeroListener), ErrExists)
}

func (s *listeners) Adds_key_event_if_none_of_the_above(t *T) {
	kk := s.fx.LL(t)
	t.FatalOn(kk.Key(tcell.KeyBS, 0, zeroListener))
	l, ok := kk.KeyListenerOf(tcell.KeyBS, 0)
	t.True(ok)
	t.Eq(fmt.Sprintf("%p", zeroListener), fmt.Sprintf("%p", l))
}

func (s *listeners) Removes_key_event_if_given_listener_nil(t *T) {
	kk := s.fx.LL(t)
	t.FatalOn(kk.Key(tcell.KeyBS, 0, zeroListener))
	t.FatalOn(kk.Key(tcell.KeyBS, 0, nil))
	t.FatalOn(kk.Key(tcell.KeyBS, 0, zeroListener))
}

func (s *listeners) Discriminates_key_events_by_mode_mask(t *T) {
	kk, shiftListener := s.fx.LL(t), func(*Env) {}
	t.FatalOn(kk.Key(tcell.KeyBS, 0, zeroListener))
	t.FatalOn(kk.Key(tcell.KeyBS, tcell.ModShift, shiftListener))
	zl, ok := kk.KeyListenerOf(tcell.KeyBS, 0)
	sl, ko := kk.KeyListenerOf(tcell.KeyBS, tcell.ModShift)
	t.True(ok && ko)
	t.Eq(fmt.Sprintf("%p", zeroListener), fmt.Sprintf("%p", zl))
	t.Eq(fmt.Sprintf("%p", shiftListener), fmt.Sprintf("%p", sl))
	t.FatalOn(kk.Key(tcell.KeyBS, 0, nil))
	_, ok = kk.KeyListenerOf(tcell.KeyBS, 0)
	t.False(ok)
	_, ok = kk.KeyListenerOf(tcell.KeyBS, tcell.ModShift)
	t.True(ok)
}

func (s *listeners) Have_a_registered_keyboard_listener(t *T) {
	kk := s.fx.LL(t)
	kk.Keyboard(func(_ *Env, r rune, k tcell.Key, mm tcell.ModMask) {})
	t.True(kk.HasKBListener())
}

func (s *listeners) Remove_a_registered_keyboard_listener(t *T) {
	kk := s.fx.LL(t)
	kk.Keyboard(func(_ *Env, r rune, k tcell.Key, mm tcell.ModMask) {})
	t.True(kk.HasKBListener())
	kk.Keyboard(nil)
	t.False(kk.HasKBListener())
}

func TestKeys(t *testing.T) {
	t.Parallel()
	Run(&listeners{}, t)
}
