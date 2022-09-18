// copyright (c) 2022 stephan lukits. all rights reserved.
// use of this source code is governed by a mit-style
// license that can be found in the license file.

package controller

import (
	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type Buttons struct {
	Suite
	Fixtures
}

func (s *Buttons) SetUp(t *T) { t.Parallel() }

func (s *Buttons) TearDown(t *T) { s.Get(t).(func())() }

func (s *Buttons) fx(t *T) (*lines.Events, *Testing) {
	return fx(t, s)
}

var dfltBttFX = []string{"[v]et=off", "[r]ace=off", "[s]tats=off", "[m]ore"}
var moreBttFX = []string{"[h]elp", "[a]bout", "[q]uit", "[b]ack"}

func (s *Buttons) Init(t *S) {
	initGolden(t)
}

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

func (s *Buttons) Switches_vet_button(t *T) {
	_, tt := s.fx(t)

	label, vw := tt.dfltButtonLabel("vet")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.dfltButtonLabel("vet")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.dfltButtonLabel("vet")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}

func (s *Buttons) Switches_race_arg(t *T) {
	_, tt := s.fx(t)

	label, vw := tt.dfltButtonLabel("race")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.dfltButtonLabel("race")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.dfltButtonLabel("race")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}

func (s *Buttons) Switches_stats_arg(t *T) {
	_, tt := s.fx(t)

	label, vw := tt.dfltButtonLabel("race")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.dfltButtonLabel("race")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.dfltButtonLabel("race")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}
