// copyright (c) 2022 stephan lukits. all rights reserved.
// use of this source code is governed by a mit-style
// license that can be found in the license file.

package controller

import (
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type Buttons struct {
	Suite
	Fixtures
}

func (s *Buttons) SetUp(t *T) { t.Parallel() }

func (s *Buttons) TearDown(t *T) {
	fx := s.Get(t)
	if fx == nil {
		return
	}
	fx.(func())()
}

func (s *Buttons) fx(t *T) (*lines.Events, *Testing) {
	return fx(t, s)
}

var dfltBttFX = []string{"[p]kgs", "[s]uites=off", "[a]rgs", "[m]ore"}
var moreBttFX = []string{"[h]elp", "[a]bout", "[q]uit", "[b]ack"}
var argsBttFX = []string{"[r]ace=off", "[v]et=off", "[s]tats=off",
	"[b]ack"}

func (s *Buttons) Switch_to_more_if_more_selected(t *T) {
	_, tt := s.fx(t)
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.ClickButton("more")
	t.SpaceMatched(tt.ButtonBar().String(), moreBttFX...)

	tt.ClickButton("back")
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.FireRune('m')
	t.SpaceMatched(tt.ButtonBar().String(), moreBttFX...)
}

func (s *Buttons) Switch_to_args_if_args_selected(t *T) {
	_, tt := s.fx(t)
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.ClickButton("args")
	t.SpaceMatched(tt.ButtonBar().String(), argsBttFX...)

	tt.ClickButton("back")
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.FireRune('a')
	t.SpaceMatched(tt.ButtonBar().String(), argsBttFX...)
}

func (s *Buttons) Switches_vet_arg(t *T) {
	_, tt := s.fx(t)

	tt.ClickButton("args")
	label, vw := tt.ArgButtonLabel("vet")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.ArgButtonLabel("vet")
	t.FatalIfNot(t.Neq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.ArgButtonLabel("vet")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}

func (s *Buttons) Switches_race_arg(t *T) {
	_, tt := s.fx(t)

	tt.ClickButton("args")
	label, vw := tt.ArgButtonLabel("race")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.ArgButtonLabel("race")
	t.FatalIfNot(t.Neq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.ArgButtonLabel("race")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}

func (s *Buttons) Switches_stats_arg(t *T) {
	_, tt := s.fx(t)

	tt.ClickButton("args")
	label, vw := tt.ArgButtonLabel("race")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.ArgButtonLabel("race")
	t.FatalIfNot(t.Neq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.ArgButtonLabel("race")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}

func TestButtons(t *testing.T) {
	t.Parallel()
	Run(&Buttons{}, t)
}
