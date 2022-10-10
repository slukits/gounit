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

func (s *Buttons) TearDown(t *T) { s.Get(t).(func())() }

func (s *Buttons) fx(t *T) (*lines.Events, *Testing) {
	return fx(t, s)
}

var switchBttFX = []string{"[v]et=off", "[r]ace=off", "[s]tats=off", "[b]ack"}
var dfltBttFX = []string{"[s]witches", "[h]elp", "[a]bout", "[q]uit"}

func (s *Buttons) Init(t *S) {
	initGolden(t)
}

func (s *Buttons) Switch_to_switches_if_switches_selected(t *T) {
	_, tt := s.fx(t)
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.ClickButton("switches")
	t.SpaceMatched(tt.ButtonBar().String(), switchBttFX...)

	tt.ClickButton("back")
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.FireRune('s')
	t.SpaceMatched(tt.ButtonBar().String(), switchBttFX...)
}

func (s *Buttons) Switches_vet_button(t *T) {
	_, tt := s.fx(t)
	tt.ClickButton("switches")

	label, vw := tt.switchButtonLabel("vet")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.switchButtonLabel("vet")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.switchButtonLabel("vet")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}

func (s *Buttons) Switches_race_button(t *T) {
	_, tt := s.fx(t)
	tt.ClickButton("switches")

	label, vw := tt.switchButtonLabel("race")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.switchButtonLabel("race")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.switchButtonLabel("race")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}

func (s *Buttons) Switches_stats_button(t *T) {
	_, tt := s.fx(t)
	tt.ClickButton("switches")

	label, vw := tt.switchButtonLabel("stats")
	t.Contains(tt.ButtonBar().String(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.switchButtonLabel("stats")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.switchButtonLabel("stats")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBar().String(), vw2)
}

func TestButtons(t *testing.T) {
	t.Parallel()
	Run(&Buttons{}, t)
}
