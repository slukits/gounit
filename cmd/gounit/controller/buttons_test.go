// copyright (c) 2022 stephan lukits. all rights reserved.
// use of this source code is governed by a mit-style
// license that can be found in the license file.

package controller

import (
	"testing"

	. "github.com/slukits/gounit"
)

type Buttons struct {
	Suite
}

func (s *Buttons) fx(t *T) *Testing {
	return fx(t)
}

var switchBttFX = []string{
	"[v]et=off", "[r]ace=off", "[s]tats=off", "[b]ack"}
var dfltBttFX = []string{"[s]witches", "[h]elp", "[a]bout", "[q]uit"}

func (s *Buttons) Init(t *S) { initGolden(t) }

func (s *Buttons) SetUp(t *T) { t.Parallel() }

func (s *Buttons) Switch_to_switches_if_switches_selected(t *T) {
	tt := s.fx(t)
	t.SpaceMatched(tt.ButtonBarCells(), dfltBttFX...)

	tt.ClickButton("switches")
	t.SpaceMatched(tt.ButtonBarCells(), switchBttFX...)

	tt.ClickButton("back")
	t.SpaceMatched(tt.ButtonBarCells(), dfltBttFX...)

	tt.FireRune('s')
	t.SpaceMatched(tt.ButtonBarCells(), switchBttFX...)
}

// type dbg struct{ Suite }
//
// func (s *dbg) fx(t *T) *Testing {
// 	return fxDBG(t)
// }
//
// func (s *dbg) Dbg(t *T) {
// 	tt := s.fx(t)
// 	t.SpaceMatched(tt.ButtonBarCells(), dfltBttFX...)
//
// 	tt.ClickButton("switches")
// 	t.SpaceMatched(tt.ButtonBarCells(), switchBttFX...)
//
// 	tt.ClickButton("back")
// 	t.SpaceMatched(tt.ButtonBarCells(), dfltBttFX...)
//
// 	tt.FireRune('s')
// 	t.SpaceMatched(tt.ButtonBarCells(), switchBttFX...)
// }
//
// func TestDBG(t *testing.T) { Run(&dbg{}, t) }

func (s *Buttons) Switches_vet_button(t *T) {
	tt := s.fx(t)
	tt.ClickButton("switches")
	t.SpaceMatched(tt.ButtonBarCells(), switchBttFX...)

	label, vw := tt.switchButtonLabel("vet")
	t.Contains(tt.ButtonBarCells(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.switchButtonLabel("vet")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBarCells(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.switchButtonLabel("vet")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBarCells(), vw2)
}

func (s *Buttons) Switches_race_button(t *T) {
	tt := s.fx(t)
	tt.ClickButton("switches")
	t.SpaceMatched(tt.ButtonBarCells(), switchBttFX...)

	label, vw := tt.switchButtonLabel("race")
	t.Contains(tt.ButtonBarCells(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.switchButtonLabel("race")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBarCells(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.switchButtonLabel("race")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBarCells(), vw2)
}

func (s *Buttons) Switches_stats_button(t *T) {
	tt := s.fx(t)
	tt.ClickButton("switches")
	t.SpaceMatched(tt.ButtonBarCells(), switchBttFX...)

	label, vw := tt.switchButtonLabel("stats")
	t.Contains(tt.ButtonBarCells(), vw)

	tt.ClickButton(label)
	label, vw2 := tt.switchButtonLabel("stats")
	t.FatalIfNot(t.Not.Eq(vw, vw2))
	t.Contains(tt.ButtonBarCells(), vw2)

	tt.ClickButton(label)
	_, vw2 = tt.switchButtonLabel("stats")
	t.FatalIfNot(t.Eq(vw, vw2))
	t.Contains(tt.ButtonBarCells(), vw2)
}

func TestButtons(t *testing.T) {
	t.Parallel()
	Run(&Buttons{}, t)
}
