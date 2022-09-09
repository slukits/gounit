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

	tt.waitFor(func() { tt.ClickButton("more") })
	t.SpaceMatched(tt.ButtonBar().String(), moreBttFX...)

	tt.waitFor(func() { tt.ClickButton("back") })
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.waitFor(func() { tt.FireRune('m') })
	t.SpaceMatched(tt.ButtonBar().String(), moreBttFX...)
}

func (s *Buttons) Switch_to_args_if_args_selected(t *T) {
	_, tt := s.fx(t)
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.waitFor(func() { tt.ClickButton("args") })
	t.SpaceMatched(tt.ButtonBar().String(), argsBttFX...)

	tt.waitFor(func() { tt.ClickButton("back") })
	t.SpaceMatched(tt.ButtonBar().String(), dfltBttFX...)

	tt.waitFor(func() { tt.FireRune('a') })
	t.SpaceMatched(tt.ButtonBar().String(), argsBttFX...)
}

func TestButtons(t *testing.T) {
	t.Parallel()
	Run(&Buttons{}, t)
}
