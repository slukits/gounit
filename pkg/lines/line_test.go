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

type TheZeroLine struct{ Suite }

func (s *TheZeroLine) Has_the_zero_type(t *T) {
	t.Eq(0, Zero.Type())
}

func (s *TheZeroLine) Does_not_accept_a_type_update(t *T) {
	t.False(Zero.SetType(42))
	t.Eq(0, Zero.Type())
}

func (s *TheZeroLine) Does_not_get_dirty(t *T) {
	t.False(Zero.Set("42").IsDirty())
}

func (s *TheZeroLine) Does_not_accept_content_setting(t *T) {
	current, stale := Zero.Set("42").Get()
	t.Eq("", current)
	t.Eq("", stale)
}

func TestTheZeroLine(t *testing.T) { Run(&TheZeroLine{}, t) }

type ALine struct {
	Suite
	fx *FX
}

type fixture struct {
	EE *Events
	TT *Testing
}

type FX struct {
	*Fixtures
	DefaultLineCount int
}

// NewFx instantiates a new concurrency save fixture set.
func NewFX() *FX {
	return &FX{
		Fixtures:         &Fixtures{},
		DefaultLineCount: 25,
	}
}

// New creates a new fixture for given test.
func (fx *FX) New(t *T) {
	ee, tt := Test(t.GoT())
	fx.Set(t, &fixture{EE: ee, TT: tt})
}

// For returns the fixture for given setting the maximum of reported
// events if given.
func (fx *FX) For(t *T, max ...int) (*Events, *Testing) {
	fix := fx.Get(t).(*fixture)
	if len(max) > 0 {
		fix.TT.SetMax(max[0])
	}
	return fix.EE, fix.TT
}

// Del removes given test's fixture and in case the fixtures Events
// instance is still listening it is quit.
func (fx *FX) Del(t *T) interface{} {
	fix, ok := fx.Fixtures.Del(t).(*fixture)
	if !ok {
		return nil
	}
	if fix.EE.IsListening() {
		fix.EE.QuitListening()
	}
	return fix
}

func (s *ALine) Init(t *I) {
	s.fx = NewFX()
}

func (s *ALine) SetUp(t *T) {
	t.Parallel()
	s.fx.New(t)
}

func (s *ALine) TearDown(t *T) {
	s.fx.Del(t)
}

func (s *ALine) Is_dirty_after_its_content_changes(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(v *Env) {
		t.False(v.LL().Line(0).Set("").IsDirty())
		t.True((v.LL().Line(0)).Set("42").IsDirty())
	})
	ee.Listen()
}

func (s *ALine) Prints_its_content_with_the_first_resize(t *T) {
	ee, tt := s.fx.For(t)
	exp := "line 0"
	ee.Resize(func(v *Env) { v.LL().Line(0).Set(exp) })
	ee.Listen()
	t.Eq(exp, fmt.Sprint(tt.LastScreen))
}

func (s *ALine) Can_have_its_type_changed(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(v *Env) {
		v.LL().Line(0).SetType(42)
		t.Eq(42, v.LL().Line(0).Type())
		v.LL().Line(0).SetType(0)
		t.Eq(42, v.LL().Line(0).Type())
	})
	ee.Listen()
}

func (s *ALine) Updates_on_screen_with_content_changing_event(t *T) {
	ee, tt := s.fx.For(t, 2)
	init, update := "line 0", "update 0"
	ee.Resize(func(v *Env) { v.LL().Line(0).Set(init) })
	ee.Rune('u', func(v *Env) { v.LL().Line(0).Set(update) })
	ee.Listen()
	t.Eq(init, tt.String())
	tt.FireRune('u')
	t.Eq(update, fmt.Sprint(tt.LastScreen))
}

func (s *ALine) Is_not_dirty_after_screen_synchronization(t *T) {
	ee, tt := s.fx.For(t, 6)
	ee.Resize(func(e *Env) {
		e.LL().Line(0).Set("line 0")
		t.True(e.LL().Line(0).IsDirty())
	})
	ee.Rune('a', func(e *Env) {
		e.LL().Line(0).Set("rune 0")
		t.True(e.LL().Line(0).IsDirty())
	})
	ee.Key(tcell.KeyUp, 0, func(v *Env) {
		v.LL().Line(0).Set("key 0")
		t.True(v.LL().Line(0).IsDirty())
	})
	t.FatalOn(ee.Update(func(e *Env) {
		t.False(e.LL().Line(0).IsDirty())
	}))
	tt.FireRune('a')
	t.FatalOn(ee.Update(func(e *Env) {
		t.False(e.LL().Line(0).IsDirty())
	}))
	tt.FireKey(tcell.KeyUp)
	t.FatalOn(ee.Update(func(e *Env) {
		t.False(e.LL().Line(0).IsDirty())
	}))
}

func (s *ALine) Pads_a_shrinking_line_with_blanks(t *T) {
	ee, tt := s.fx.For(t, 2)
	long, short := "a longer line", "short line"
	ee.Resize(func(e *Env) { e.LL().Line(0).Set(long) })
	ee.Rune('a', func(e *Env) { e.LL().Line(0).Set(short) })
	ee.Listen()
	t.Eq(long, tt.String())
	tt.FireRune('a')
	t.Eq(short, fmt.Sprint(tt.LastScreen))
}

func TestALine(t *testing.T) {
	t.Parallel()
	Run(&ALine{}, t)
}
