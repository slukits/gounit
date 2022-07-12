// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
	"github.com/slukits/gounit/pkg/lines/testdata/fx"
)

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
	s.fx.Set(t, fx.NewView(t))
}

func (s *ALine) TearDown(t *T) {
	s.fx.Del(t)
}

func (s *ALine) Is_dirty_after_its_content_changes(t *T) {
	v := s.fx.View(t)
	v.Register.Resize(func(v *lines.View) {
		t.False(v.Line(0).Set("").IsDirty())
		t.True((v.Line(0)).Set("42").IsDirty())
	})
	v.Listen()
}

func (s *ALine) Prints_its_content_with_the_first_resize(t *T) {
	v, exp := s.fx.View(t), "line 0"
	v.Register.Resize(func(v *lines.View) { v.Line(0).Set(exp) })
	v.Listen()
	t.Eq(exp, v.LastScreen)
}

func (s *ALine) Updates_on_screen_with_content_changing_event(t *T) {
	v, init, update := s.fx.View(t, 1), "line 0", "update 0"
	v.Register.Resize(func(v *lines.View) { v.Line(0).Set(init) })
	v.Register.Rune(func(v *lines.View) { v.Line(0).Set(update) }, 'u')
	go v.Listen()
	<-v.NextEventProcessed
	t.Eq(init, v.String())
	<-v.FireRuneEvent('u')
	t.Eq(update, v.LastScreen)
}

func TestALine(t *testing.T) {
	t.Parallel()
	Run(&ALine{}, t)
}

type DBG struct{ Suite }

func TestDBG(t *testing.T) { Run(&DBG{}, t) }
