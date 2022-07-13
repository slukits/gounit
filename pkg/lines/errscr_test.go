// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"strings"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
	"github.com/slukits/gounit/pkg/lines/testdata/fx"
)

type AErrorScreen struct {
	Suite
	fx FX
}

func (s *AErrorScreen) Init(t *I) {
	s.fx.Fixtures = &Fixtures{}
	s.fx.DefaultLineCount = 25
}

func (s *AErrorScreen) SetUp(t *T) {
	t.Parallel()
	s.fx.Set(t, fx.New(t))
}

func (s *AErrorScreen) TearDown(t *T) { s.fx.Del(t) }

func (s *AErrorScreen) Is_dirty_after_updated(t *T) {
	rg := s.fx.Reg(t)
	rg.Resize(func(v *lines.View) {
		v.ErrScreen().Set("")
		t.False(v.ErrScreen().IsDirty())
		v.ErrScreen().Set("22")
		t.True(v.ErrScreen().IsDirty())
	})
	rg.Listen()
}

// TODO: add formatting directives to an error-screen, e.g. "centered"
func (s *AErrorScreen) A_100_percent(t *T) {
	rg := s.fx.Reg(t)
	rg.Resize(func(v *lines.View) {
		v.ErrScreen().Set(strings.Repeat("a", 108))
		v.ErrScreen().Active = true
	})
	rg.Listen()
}

func TestAErrorScreen(t *testing.T) {
	t.Parallel()
	Run(&AErrorScreen{}, t)
}
