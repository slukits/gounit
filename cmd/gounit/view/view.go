// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"errors"
	"fmt"

	"github.com/slukits/lines"
)

// An Initer implementation initializes a new view provided to the
// [View.New] constructor and it is provided with the functionality to
// manipulate view, i.e. the screen content.
type Initer interface {

	// Message returns the message bar's default content and is provided
	// by a View with a function to update or reset the message bar's
	// content.  Calling update with the empty string resets the message
	// bar's content.
	Message(update func(string)) string

	// Status returns the statusbar's default content and is provided by
	// a View with a function to update or reset the statusbar's
	// content.  Calling update with the empty string resets the
	// statusbar's content.
	Status(update func(string)) string

	// Main returns an initial content a View's main area.
	Main() string

	// ForButton implementation is provided by a View with a callback
	// function which may be used to initialize the View's button bar.
	// Such a callback may be called with a [View.ButtonUpdater]
	// callback which provides to the Initer implementation a function
	// to update a button definition.  A button definition fails if
	// ambiguous button labels are provided.
	ForButton(func(ButtonDef, ButtonUpdater) error)
}

// view implements the lines Componenter interface hence an instance of
// it can be used to initialize a lines terminal ui.  Note the view
// package is designed in a way that there shouldn't be done anything
// else with a view instance but initializing a lines Events instance.
// A view instance may be modified by the provided functions to an
// Initer implementation.
type view struct {
	lines.Component
	lines.Stacking
	ee          *lines.Events
	runeButtons map[rune]*button
}

// New uses provided information of given Initer i implementation to
// initialize a new returned view instance.  In turn the Initer
// implementation is provided with the functionality to modify created
// view instance.  New's return value implements the lines.Componenter
// interface and should be only ever used to initialize a lines Events
// instance, e.g.:
//
//	lines.New(view.New(i))
func New(i Initer) *view {
	new := &view{}
	new.CC = append(new.CC, &messageBar{
		dflt: i.Message(new.updateMessageBar)})
	new.CC = append(new.CC, &main{dflt: i.Main()})
	new.CC = append(new.CC, &statusBar{
		dflt: i.Status(new.updateStatusBar)})
	initButtons(i, new)
	return new
}

func initButtons(i Initer, v *view) *buttonBar {
	bb := &buttonBar{}
	v.CC = append(v.CC, bb)
	i.ForButton(func(bd ButtonDef, bu ButtonUpdater) error {

		if err := v.validateButtonDef(bd, nil); err != nil {
			return err
		}

		// add button for valid button definition
		bb.bb = append(bb.bb, &button{
			label: bd.Label, listener: bd.Listener})
		if bd.Rune != 0 {
			v.addRune(bd.Rune, bb.bb[len(bb.bb)-1])
		}
		if bu != nil {
			bu(v.updateButtonClosure(bb.bb[len(bb.bb)-1]))
		}
		return nil
	})
	return bb
}

func (v *view) OnInit(e *lines.Env) {
	v.ee = e.EE
}

func (v *view) OnRune(_ *lines.Env, r rune) {
	b, ok := v.runeButtons[r]
	if !ok || b.listener == nil {
		return
	}
	b.listener(b.label)
}

func (v *view) updateMessageBar(s string) {
	v.ee.Update(v.CC[0], s, nil)
}

func (v *view) updateStatusBar(s string) {
	v.ee.Update(v.CC[2], s, nil)
}

func (v *view) addRune(r rune, b *button) {
	if v.runeButtons == nil {
		v.runeButtons = map[rune]*button{}
	}
	v.runeButtons[r] = b
}

func (v *view) updateButtonClosure(btt *button) func(ButtonDef) error {
	return func(bd ButtonDef) error {

		if err := v.validateButtonDef(bd, btt); err != nil {
			return err
		}
		if err := v.ee.Update(btt, &bd, nil); err != nil {
			return err
		}
		if v.runeButtons[bd.Rune] == btt {
			return nil
		}
		for r, b := range v.runeButtons { // NOTE: not concurrency save
			if btt != b {
				continue
			}
			delete(v.runeButtons, r)
			if bd.Rune == 0 {
				break
			}
			v.runeButtons[bd.Rune] = b
		}
		return nil
	}
}

// ErrButtonLabelAmbiguity is returned during a view's initialization or
// from a button update iff a button should be created/updated to a
// label which is already used by an other button.
var ErrButtonLabelAmbiguity = errors.New(
	"view: define button: ambiguous label: ")

// ErrButtonRuneAmbiguity is returned during a view's initialization or
// from a button update iff a button should be created/updated to a
// button rune which is already used by an other button.
var ErrButtonRuneAmbiguity = errors.New(
	"view: define button: ambiguous rune: ")

func (v *view) validateButtonDef(bd ButtonDef, btt *button) error {
	if b, ok := v.runeButtons[bd.Rune]; ok && b != btt {
		return fmt.Errorf(
			"%w%c", ErrButtonRuneAmbiguity, bd.Rune)
	}
	if bd.Label == "" {
		return nil
	}
	for _, b := range v.CC[3].(*buttonBar).bb {
		if btt != nil && btt == b {
			continue
		}
		if b.label == bd.Label {
			return fmt.Errorf(
				"%w%s", ErrButtonLabelAmbiguity, bd.Label)
		}
	}
	return nil
}
