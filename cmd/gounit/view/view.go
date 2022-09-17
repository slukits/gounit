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
	|      [v]et=off      [r]ace=off      [s]tats=off      [m]ore      |
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
	Status(update func(Statuser))

	// Reporting is by initializing view provided with an update
	// function for the reporting component's lines and expects an initial
	// content string as well as a listener which is notified if a user
	// selects a line.
	Reporting(update func(Reporter)) Reporter

	// Buttons is by initializing view provided with an update function
	// for buttons and expects a Buttoner implementation with the
	// initial button definitions.
	Buttons(update func(Buttoner)) Buttoner
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

	// PackageLine classifies a reported line as package-line.  Note
	// only package or suite lines are selectable.
	PackageLine

	// TestLine classifies a reported line as go-test-line.  Note a
	// go-test-line is not selectable.
	TestLine

	// SuiteLine classifies a reported line as suite-line.  Note only
	// package or suite lines are selectable.
	SuiteLine

	// SuiteTestLine classifies a reported line as suit-test-line.  Note
	// a suit-test-line is not selectable.
	SuiteTestLine

	// ZeroLineMode indicates no other than default formattings for a
	// line of a reporting component.
	ZeroLineMod LineMask = 0
)

// View implements the lines Componenter interface hence an instance of
// it can be used to initialize a lines terminal ui.  Note the View
// package is designed in a way that there shouldn't be done anything
// else with a View instance but initializing a lines Events instance.
// A View instance may be modified by the provided functions to an
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
	r := i.Reporting(new.updateLines)
	new.CC = append(new.CC, &report{rr: []Reporter{r}})
	new.CC = append(new.CC, &statusBar{})
	i.Status(new.updateStatusBar)
	initButtons(i, new)
	return new
}

func initButtons(i Initer, v *view) {
	bb := &buttonBar{}
	v.CC = append(v.CC, bb) // necessary for button-def validation
	bbDef := i.Buttons(v.updateButtons)
	if err := v.validateButtoner(bbDef, true); err != nil {
		return
	}
	bb.init(&buttonsUpdate{bb: bbDef, setRune: v.setRune})
}

func (v *view) OnInit(e *lines.Env) {
	v.ee = e.EE
	if err := e.EE.MoveFocus(v.reporting()); err != nil {
		v.fatal(fmt.Sprintf("gounit: view: move focus: %v", err))
	}
	width, _ := e.ScreenSize()
	if width > 80 {
		v.Dim().SetWidth(80)
	}
}

func (v *view) reporting() lines.Componenter { return v.CC[1] }

func (v *view) OnLayout(e *lines.Env) {
	width, _ := e.ScreenSize()
	if width > 80 {
		v.Dim().SetWidth(80)
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

func (v *view) updateStatusBar(upd Statuser) {
	if err := v.ee.Update(v.CC[2], upd, nil); err != nil {
		v.fatal(fmt.Sprintf("gounit: view: update: statusbar: %v", err))
	}
}

func (v *view) updateLines(l Reporter) {
	if err := v.ee.Update(v.CC[1], l, nil); err != nil {
		v.fatal(fmt.Sprintf("gounit: view: update: lines: %v", err))
	}
}

type buttonsUpdate struct {
	bb      Buttoner
	setRune func(rune, string, *button)
}

func (v *view) updateButtons(bb Buttoner) {
	if bb.Replace() {
		if err := v.validateButtoner(bb, true); err != nil {
			return
		}
		v.runeButtons = map[rune]*button{}
		v.ee.Update(v.CC[3], &buttonsUpdate{
			bb: bb, setRune: v.setRune}, nil)
		return
	}
	if err := v.validateButtoner(bb, false); err != nil {
		return
	}
	v.ee.Update(v.CC[3], &buttonsUpdate{
		bb: bb, setRune: v.setRune}, nil)
}

func (v *view) setRune(r rune, label string, btt *button) {
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
	if v.runeButtons == nil {
		v.runeButtons = map[rune]*button{}
	}
	v.runeButtons[r] = btt
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

var ErrLabelMustNotBeZero = errors.New(
	"view: define button: label must not be zero")

func (v *view) validateButtoner(bb Buttoner, init bool) error {
	if init {
		return v.validateInitButtoner(bb)
	}

	ll := map[string]*button{}
	for _, b := range v.CC[3].(*buttonBar).bb {
		ll[b.label] = b
	}
	rr := map[rune]*button{}
	for r, b := range v.runeButtons {
		rr[r] = b
	}

	// check for ambiguities amongst new buttons
	var err error
	bb.ForNew(func(bd ButtonDef) error {
		if err != nil {
			return err
		}
		if bd.Label == "" {
			err = ErrLabelMustNotBeZero
			return err
		}
		if _, ok := ll[bd.Label]; !ok {
			err = fmt.Errorf("%w%s", ErrButtonLabelAmbiguity, bd.Label)
			return err
		}
		ll[bd.Label] = nil
		if _, ok := rr[bd.Rune]; !ok {
			err = fmt.Errorf("%w%c", ErrButtonRuneAmbiguity, bd.Rune)
			return err
		}
		if bd.Rune == 0 {
			return nil
		}
		rr[bd.Rune] = nil
		return nil
	})
	if err != nil {
		return err
	}

	// check for not found buttons and removal of buttons
	bb.ForUpdate(func(label string, bd ButtonDef) error {
		if err != nil {
			return err
		}
		if b, ok := ll[label]; !ok || b == nil {
			err = fmt.Errorf("%w%s", ErrButtonNotFound, label)
			return err
		}
		if bd.Label == "" {
			b := ll[label]
			for r, rb := range rr {
				if b != rb {
					continue
				}
				delete(rr, r)
				break
			}
			delete(ll, label)
		}
		return nil
	})
	if err != nil {
		return err
	}

	// check ambiguities in updated not deleted buttons
	bb.ForUpdate(func(label string, bd ButtonDef) error {
		if b, ok := ll[bd.Label]; ok && (b == nil || b.label != label) {
			err = fmt.Errorf("%w%s", ErrButtonLabelAmbiguity, bd.Label)
			return err
		}
		if b, ok := rr[bd.Rune]; ok && b != ll[label] {
			err = fmt.Errorf("%w%c", ErrButtonRuneAmbiguity, bd.Rune)
			return err
		}
		return nil
	})

	return err
}

func (v *view) validateInitButtoner(bb Buttoner) error {
	ll := map[string]bool{}
	rr := map[rune]bool{}
	var err error
	bb.ForNew(func(bd ButtonDef) error {
		if err != nil {
			return err
		}
		if bd.Label == "" {
			err = ErrLabelMustNotBeZero
			return err
		}
		if ll[bd.Label] {
			err = fmt.Errorf("%w%s", ErrButtonLabelAmbiguity, bd.Label)
			return err
		}
		ll[bd.Label] = true
		if rr[bd.Rune] {
			err = fmt.Errorf("%w%c", ErrButtonRuneAmbiguity, bd.Rune)
			return err
		}
		if bd.Rune == 0 {
			return nil
		}
		rr[bd.Rune] = true
		return nil
	})

	return err
}
