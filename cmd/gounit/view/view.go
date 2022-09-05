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
	|  [p]kgs   [s]uites   [v]et=off   [r]ace=off   [s]ats=on  [q]uit  |
	+------------------------------------------------------------------+
*/
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

	// Fatal provides the function to which a view reports fatal
	// view-errors to.
	Fatal() func(...interface{})

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

	// Report returns an initial content a View's main area and is
	// provided by a View with a function to update line-listeners and
	// an other to update lines.
	Report(LLUpdater, LinesUpdater) string

	// ForButton implementation is provided by a View with a callback
	// function which may be used to initialize the View's button bar.
	// Such a callback may be called with a [View.ButtonUpdater]
	// callback which provides to the Initer implementation a function
	// to update a button definition.  A button definition fails if
	// ambiguous button labels are provided.
	ForButton(func(ButtonDef, ButtonUpdater) error)
}

// LLMod  values are used to describe to a main lines-listener what kind
// of event is reported.
type LLMod uint8

const (
	// Context is reported iff a line is clicked with the "right" mouse
	// key.
	Context LLMod = 1 << iota
	// Default is reported iff a line is clicked or other wise selected
	// without special conditions (its still undefined what special
	// conditions are).
	Default LLMod = 0
)

// LLUpdater function is provided  to an Initer implementation to update
// lines listeners.  A lines listener (function) is informed if a particular line
// was selected by the user, e.g. clicked.
type LLUpdater func(func(idx int, mod LLMod))

// A Liner provides line-updates for the main area.
type Liner interface {
	// Clearing returns true iff all remaining lines which are not set
	// by Liner.For of the main component should be cleared.
	Clearing() bool
	// calls given function for each line back.
	For(line func(idx uint, content string))
}

// UpdateLines function is provided to an Initer implementation.  It
// expects an callback function which is defining a line.
type LinesUpdater func(Liner)

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
func New(i Initer) *view {
	new := &view{fatal: i.Fatal()}
	new.CC = append(new.CC, &messageBar{
		dflt: i.Message(new.updateMessageBar)})
	new.CC = append(new.CC, &report{dflt: i.Report(
		new.updateLineListener, new.updateLines)})
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

func (v *view) updateLineListener(ll func(int, LLMod)) {
	if err := v.ee.Update(v.CC[1], ll, nil); err != nil {
		v.fatal(fmt.Sprintf(
			"gounit: view: update: line-listener: %v", err))
	}
}

func (v *view) updateLines(l Liner) {
	linesUpdate := map[int]string{}
	l.For(func(idx uint, content string) {
		linesUpdate[int(idx)] = content
	})
	if len(linesUpdate) == 0 {
		return
	}
	if l.Clearing() {
		linesUpdate[-1] = "clear"
	}
	if err := v.ee.Update(v.CC[1], linesUpdate, nil); err != nil {
		v.fatal(fmt.Sprintf("gounit: view: update: lines: %v", err))
		panic(err)
	}
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
