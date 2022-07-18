// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"strings"
	"testing"

	. "github.com/slukits/gounit"
)

type AnEnvErr struct {
	Suite
	fx *FX
}

func (s *AnEnvErr) Init(t *I) { s.fx = NewFX() }

func (s *AnEnvErr) SetUp(t *T) {
	t.Parallel()
	s.fx.New(t)
}

func (s *AnEnvErr) TearDown(t *T) { s.fx.Del(t) }

func (s *AnEnvErr) Is_dirty_after_updated(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(e *Env) {
		e.Err().Set("")
		t.False(e.Err().IsDirty())
		e.Err().Set("22")
		t.True(e.Err().IsDirty())
	})
	ee.Listen()
}

// TODO: add formatting directives to an error-screen, e.g. "centered"
func (s *AnEnvErr) A_100_percent(t *T) {
	ee, _ := s.fx.For(t)
	ee.Resize(func(v *Env) {
		v.Err().Set(strings.Repeat("a", 108))
		v.Err().Active = true
	})
	ee.Listen()
}

func TestAErrorScreen(t *testing.T) {
	t.Parallel()
	Run(&AnEnvErr{}, t)
}
