// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package view utilizes the github.com/slukits/lines package to provide
gounit's terminal user interface (note: the actual ui has no frames)

	+------------------------------------------------------------------+
	|                                                                  |
	| github.com/slukits/gounit: cmd/gounit                            |
	|                                                                  |
	+------------------------------------------------------------------+
	|                                                                  |
	| view: 21/0 243ms, 10/5, 1454/684 340                             |
	|                                                                  |
	| A new view displays initially given (4/0 8ms)                    |
	|     message                                                      |
	|     status                                                       |
	|     main_info                                                    |
	|     buttons                                                      |
	|                                                                  |
	| A view 17/0 13ms                                                 |
	|                                                                  |
	| ...                                                              |
	|                                                                  |
	+------------------------------------------------------------------+
	|                                                                  |
	| packages: n; tests: t/f; stat: c/t d                             |
	+------------------------------------------------------------------+
	|     [p]kgs     [s]uites     se[t]tings     [h]elp     [q]uit     |
	+------------------------------------------------------------------+
*/
package view

import (
	"errors"
	"fmt"

	"github.com/slukits/lines"
)

// An Initer implementation initializes a new view and receives from
// initialized view function for updating view components.  See New.
type Initer interface {

	// Fatal is expected to provide a function to which initializing
	// view can report fatal view-errors to.
	Fatal() func(...interface{})

	// Message is by initializing view provided with an update function
	// for the message bar's content and expects an initial content.
	// Calling update with the empty string resets the message bar's
	// content.
	Message(update func(string)) string

	// Status is by initializing view provided with an update function
	// for the status bar's content and expects an initial content.
	// Calling update with the empty string resets the statusbar's
	// content.
	Status(update func(string)) string

	// Reporting is by initializing view provided with an update
	// function for the reporting component's lines and expects an initial
	// content string as well as a listener which is notified if a user
	// selects a line.
	Reporting(ReportingUpd) (string, ReportingLst)

	// Buttons is by initializing view provided with an update function
	// for buttons and expects a function for initial button definitions
	// and a listener function which is notified if a button was
	// selected.
	Buttons(_ ButtonUpd, bb func(ButtonDef) error) ButtonLst
}

// LineMask  values are used to describe to a reporting component how a
// particular line should be displayed.
type LineMask uint8

const (

	// Failed sets error formattings for a reporting component's line
	// like a red background and a white foreground.
	Failed LineMask = 1 << iota

	// Passed sets a lines as the "green bar", i.e. a green background
	// and a black foreground.
	Passed

	// ZeroLineMode indicates no other than default formattings for a
	// line of a reporting component.
	ZeroLineMod LineMask = 0
)

// LstUpdater function is provided  to an Initer implementation to
// update the lines listener.  A lines listener (function) is informed
// if a particular line was selected by the user, e.g. clicked.
type LstUpdater func(func(idx int))

// A Liner provides line-updates for the main area.
type Liner interface {

	// Clearing returns true iff all remaining lines which are not set
	// by Liner.For of the main component should be cleared.
	Clearing() bool

	// For is provided with the reporting component instance and a
	// callback function which must be called for each line which should
	// be updated.  If Clearing all other lines of reporting component
	// are reset to zero.  For each updated line Mask is called for
	// optional formatting information.
	For(_ lines.Componenter, line func(idx uint, content string))

	// Mask may provide for an updated line additional formatting
	// information like "Failed" or "Passed".
	Mask(idx uint) LineMask
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
	fatal       func(...interface{})
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
//
// or a testing instance
//
//	func (s *MySuite) My_suite_test(t *T) {
//		tt := view.Test{t, view.New(i)}
//		ee, lt := lines.Test(tt)
//	}
func New(i Initer) *view {
	new := &view{fatal: i.Fatal()}
	new.CC = append(new.CC, &messageBar{
		dflt: i.Message(new.updateMessageBar)})
	dflt, lst := i.Reporting(new.updateLines)
	new.CC = append(new.CC, &report{dflt: dflt, listener: lst})
	new.CC = append(new.CC, &statusBar{
		dflt: i.Status(new.updateStatusBar)})
	initButtons(i, new)
	return new
}

func initButtons(i Initer, v *view) *buttonBar {
	bb := &buttonBar{}
	v.CC = append(v.CC, bb) // necessary for button-def validation
	lst := i.Buttons(v.updateButton, func(bd ButtonDef) error {
		if err := v.validateButtonDef("", bd); err != nil {
			return err
		}
		bb.append(bd)
		if bd.Rune != 0 && bd.Label != "" {
			v.addRune(bd.Rune, bb.bb[len(bb.bb)-1])
		}
		return nil
	})
	bb.setListener(lst)
	return bb
}

func (v *view) OnInit(e *lines.Env) {
	v.ee = e.EE
	if err := e.EE.MoveFocus(v.CC[1]); err != nil {
		v.fatal(fmt.Sprintf("gounit: view: move focus: %v", err))
	}
}

func (v *view) OnRune(_ *lines.Env, r rune) {
	b, ok := v.runeButtons[r]
	if !ok || b.listener == nil {
		return
	}
	b.listener(b.label)
}

func (v *view) updateMessageBar(s string) {
	if err := v.ee.Update(v.CC[0], s, nil); err != nil {
		v.fatal(fmt.Sprintf(
			"gounit: view update: message-bar: %v", err))
	}
}

func (v *view) updateStatusBar(s string) {
	if err := v.ee.Update(v.CC[2], s, nil); err != nil {
		v.fatal(fmt.Sprintf("gounit: view: update: statusbar: %v", err))
	}
}

func (v *view) updateLines(l Liner) {
	if err := v.ee.Update(v.CC[1], l, nil); err != nil {
		v.fatal(fmt.Sprintf("gounit: view: update: lines: %v", err))
	}
}

func (v *view) addRune(r rune, b *button) {
	if v.runeButtons == nil {
		v.runeButtons = map[rune]*button{}
	}
	v.runeButtons[r] = b
}

type buttonUpdate struct {
	label   string
	def     ButtonDef
	setRune func(rune, *button)
}

func (v *view) updateButton(label string, def ButtonDef) error {
	if err := v.validateButtonDef(label, def); err != nil {
		return err
	}
	err := v.ee.Update(v.CC[3], &buttonUpdate{label: label, def: def,
		setRune: func(r rune, btt *button) {
			if btt == nil { // button deleted
				for r, b := range v.runeButtons {
					if b.label != label {
						continue
					}
					delete(v.runeButtons, r)
					break
				}
				return
			}
			if v.runeButtons[r] == btt { // rune not updated
				return
			}
			for r, b := range v.runeButtons { // rune updated
				if btt != b {
					continue
				}
				delete(v.runeButtons, r)
			}
			if r == 0 { // rune deleted
				return
			}
			v.runeButtons[r] = btt
		}}, nil)
	return err
}

// ErrButtonNotFound is returned if a button with given label is
// requested for update and no button with that label can be found.
var ErrButtonNotFound = errors.New(
	"view: update button: not found: ")

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

func (v *view) validateButtonDef(
	label string, bd ButtonDef,
) error {

	var btt *button
	if label != "" {
		for _, b := range v.CC[3].(*buttonBar).bb {
			if b.label != label {
				continue
			}
			btt = b
		}
		if btt == nil {
			return fmt.Errorf("%w%s", ErrButtonNotFound, label)
		}
	}
	if b, ok := v.runeButtons[bd.Rune]; ok && b != btt {
		return fmt.Errorf(
			"%w%c", ErrButtonRuneAmbiguity, bd.Rune)
	}
	for _, b := range v.CC[3].(*buttonBar).bb {
		if b.label == label {
			continue
		}
		if b.label == bd.Label {
			return fmt.Errorf(
				"%w%s", ErrButtonLabelAmbiguity, bd.Label)
		}
	}
	return nil
}
