// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

type TheZeroLine struct{ Suite }

func (s *TheZeroLine) Has_the_zero_type(t *T) {
	t.Eq(0, lines.Zero.Type())
}

func (s *TheZeroLine) Does_not_accept_a_type_update(t *T) {
	t.False(lines.Zero.SetType(42))
	t.Eq(0, lines.Zero.Type())
}

func (s *TheZeroLine) Does_not_get_dirty(t *T) {
	t.False(lines.Zero.Set("42").IsDirty())
}

func (s *TheZeroLine) Does_not_accept_content_setting(t *T) {
	current, stale := lines.Zero.Set("42").Get()
	t.Eq("", current)
	t.Eq("", stale)
}

func TestTheZeroLine(t *testing.T) { Run(&TheZeroLine{}, t) }

type ALine struct {
	Suite
	fx FX
}

func (s *ALine) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *ALine) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, New(t))
}

func (s *ALine) TearDown(t *T) {
	s.fx.Del(t)
}

func (s *ALine) Is_dirty_after_its_content_changes(t *T) {
	rg := s.fx.Reg(t)
	rg.Resize(func(v *lines.View) {
		t.False(v.LL().Line(0).Set("").IsDirty())
		t.True((v.LL().Line(0)).Set("42").IsDirty())
	})
	rg.Listen()
}

func (s *ALine) Prints_its_content_with_the_first_resize(t *T) {
	rg, exp := s.fx.Reg(t), "line 0"
	rg.Resize(func(v *lines.View) { v.LL().Line(0).Set(exp) })
	rg.Listen()
	t.Eq(exp, rg.LastScreen)
}

func (s *ALine) Can_have_its_type_changed(t *T) {
	rg := s.fx.Reg(t)
	rg.Resize(func(v *lines.View) {
		v.LL().Line(0).SetType(42)
		t.Eq(42, v.LL().Line(0).Type())
		v.LL().Line(0).SetType(0)
		t.Eq(42, v.LL().Line(0).Type())
	})
	rg.Listen()
}

func (s *ALine) Updates_on_screen_with_content_changing_event(t *T) {
	rg, init, update := s.fx.Reg(t, 1), "line 0", "update 0"
	rg.Resize(func(v *lines.View) { v.LL().Line(0).Set(init) })
	rg.Rune('u', func(v *lines.View) { v.LL().Line(0).Set(update) })
	rg.Listen()
	t.Eq(init, rg.String())
	rg.FireRuneEvent('u')
	t.Eq(update, rg.LastScreen)
}

func (s *ALine) Is_not_dirty_after_screen_synchronization(t *T) {
	rg := s.fx.Reg(t, 5)
	rg.Resize(func(v *lines.View) {
		v.LL().Line(0).Set("line 0")
		t.True(v.LL().Line(0).IsDirty())
	})
	rg.Rune('a', func(v *lines.View) {
		v.LL().Line(0).Set("rune 0")
		t.True(v.LL().Line(0).IsDirty())
	})
	rg.Key(tcell.KeyUp, 0, func(v *lines.View) {
		v.LL().Line(0).Set("key 0")
		t.True(v.LL().Line(0).IsDirty())
	})
	rg.Listen()
	err := rg.Update(
		func(v *lines.View) { t.False(v.LL().Line(0).IsDirty()) })
	t.FatalOn(err)
	rg.FireRuneEvent('a')
	err = rg.Update(
		func(v *lines.View) { t.False(v.LL().Line(0).IsDirty()) })
	t.FatalOn(err)
	rg.FireKeyEvent(tcell.KeyUp)
	err = rg.Update(
		func(v *lines.View) { t.False(v.LL().Line(0).IsDirty()) })
	t.FatalOn(err)
}

func (s *ALine) Pads_a_shrinking_line_with_blanks(t *T) {
	rg, long, short := s.fx.Reg(t, 1), "a longer line", "short line"
	rg.Resize(func(v *lines.View) { v.LL().Line(0).Set(long) })
	rg.Rune('a', func(v *lines.View) { v.LL().Line(0).Set(short) })
	rg.Listen()
	t.Eq(long, rg.String())
	rg.FireRuneEvent('a')
	t.Eq(short, rg.LastScreen)
}

func TestALine(t *testing.T) {
	t.Parallel()
	Run(&ALine{}, t)
}
