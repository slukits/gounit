// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package lines provides a simple terminal-UI where the UI is
// interpreted as an ordered set of lines.  Reported events are
//
// - up: user pressed k or cursor-up
//
// - down: user pressed j or cursor-down
//
// - enter: user pressed enter
//
// - escape: user pressed enter
//
// - quit: user pressed q, ctrl-d or ctrl-c
//
// - resize: for initial layout or user resizes the terminal window.
//
// Listeners are registered by setting corresponding properties of the
// *Listeners* property.  Calling the *Listen*-Method starts the
// event-loop with the mandatory resize-event and blocks until the quit
// event is received.  A typical example:
//
//     lv := lines.NewView()
//     lv.Listeners.Up = func(v *lines.View, mm lines.Modifiers) {
//         v.Get(0).Set("received up event")
//     }
//     lv.Listeners.Resize = func(v *lines.View) {
//         v.For(func(l *lines.Line) {
//             l.Set(fmt.Sprintf("line %d", l.Idx+1))
//         })
//         lv.Get(0).Set(fmt.Sprintf("have %d lines", v.Len()))
//     }
//     // ... register other event-listeners ...
//     lv.Listen() // block until quit
package lines

import (
	"errors"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/ints"
)

// View provides a line-based terminal user-interface.  A zero view is
// not ready to use.  *NewView* and *NewSim* create and initialize a new
// view whereas the later creates a view with tcell's simulation-screen
// for testing.  Typically you would set a new view's event-listeners
// followed by a (blocking) call of *Listen* starting the event-loop.
// Calling *Quit* stops the event-loop and releases resources.
type View struct {
	lib      tcell.Screen
	ll       []*Line
	evtRunes string
	evtKeys  ints.Set

	// Listeners holds a view's listeners which are informed about
	// occurring events. (see type *Listeners*).  Changing this property
	// after starting the event-loop will likely break event-handling.
	Listeners

	// Triggers associates rune or special-key inputs with view-events.
	// See type *Triggers* for its defaults.  Changing this property
	// after starting the event-loop will brake event-handling.
	Triggers
}

// NewView returns a new View instance or nil and an error in case
// tcell's screen-creation or its initialization fails.
func NewView() (*View, error) {
	lib, err := screenFactory.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := lib.Init(); err != nil {
		return nil, err
	}
	return &View{lib: lib, Triggers: defaultTrigger}, nil
}

// NewSim returns a new View instance wrapping tcell's simulation
// screen for testing purposes.  Since the wrapped tcell screen is
// private it is returned as well to facilitate desired mock-ups.
// NewSim fails iff tcell's screen-initialization fails.
func NewSim() (*View, tcell.SimulationScreen, error) {
	lib := screenFactory.NewSimulationScreen("")
	if err := lib.Init(); err != nil {
		return nil, nil, err
	}
	return &View{lib: lib, Triggers: defaultTrigger}, lib, nil
}

// Len returns the number of lines of a terminal screen.  Note len of
// the simulation screen defaults to 25.
func (v *View) Len() int {
	_, h := v.lib.Size()
	return h
}

// For calls ascending ordered for each line back.
func (v *View) For(cb func(*Line)) {
	for _, l := range v.ll {
		cb(l)
	}
}

// Listen starts the view's event-loop listening for user-input and
// blocks until a quit-event is received or *Quit* is called.
func (v *View) Listen() error {

	v.setupEventTriggers()

	// BUG the following four lines needed to go into the resize event
	// v.lib.Clear()
	// if v.Listeners.Resize != nil {
	//     v.Listeners.Resize(v)
	// }

	for {

		// Poll event (blocking as I understand)
		ev := v.lib.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			v.lib.Clear()
			v.ensureLines()
			if v.Listeners.Resize != nil {
				v.Listeners.Resize(v)
			}
			v.lib.Sync()
		case *tcell.EventKey:
			if ev.Key() == tcell.KeyRune {
				if strings.IndexRune(v.evtRunes, ev.Rune()) < 0 {
					continue
				}
				if err := v.triggerRuneEvent(ev.Rune()); err != nil {
					return v.err(err)
				}
			}
		default:
			return nil
		}
	}
}

func (v *View) err(e error) error {
	v.Quit()
	return e
}

func (v *View) setupEventTriggers() {
	v.evtRunes = string(v.Triggers.QuitRunes)
	v.evtKeys.Add(toInt(v.Triggers.QuitKeys)...)
}

var EventRunErr = errors.New("registered runes changed")

func (v *View) triggerRuneEvent(r rune) error {
	switch v.Triggers.resolveRuneEventType(r) {
	case quit:
		if v.Listeners.Quit != nil {
			v.Listeners.Quit()
			v.Quit()
		}
	case none:
		return EventRunErr
	}
	return nil
}

// ensureLines adapts after a resize event the view's lines-slice.
func (v *View) ensureLines() {
	if len(v.ll) == v.Len() {
		return
	}
	if len(v.ll) > v.Len() {
		v.ll = v.ll[:v.Len()]
		return
	}
	for len(v.ll) < v.Len() {
		v.ll = append(v.ll, &Line{})
	}
}

// Quit the event-loop.
func (v *View) Quit() {
	v.lib.Fini()
}

// Listeners implements a View-instance's Listeners property storing the
// event-listeners for events provided by a View-instance.  NOTE
// updating event-listeners after the event-loop has started  may lead
// to unexpected behavior especially if a test-fixture View-instance is
// used.
type Listeners struct {

	// Resize listener is called if an resize event happens.  After
	// starting a view's event loop an initial resize event is
	// mandatorily emitted.
	Resize func(*View)

	// Quit listener is called if an quit event occurred.  A quit event
	// occurs iff on of a view's quit-event triggers was received (see
	// *View.Triggers*).
	Quit func()
}

var defaultTrigger = Triggers{
	QuitRunes: []rune{'q'},
	QuitKeys:  []tcell.Key{tcell.KeyCtrlC, tcell.KeyCtrlD},
}

type evtType int8

const (
	none evtType = iota
	quit
)

// Triggers defines which rune/special-key inputs trigger which events.
// A zero value has not runes or special-keys set.  Note no checks are
// performed if different trigger-sets have the same runes/keys.  In this
// case it is undefined which event is triggered.
type Triggers struct {

	// QuitRunes are the rune-inputs which trigger a quit event. They
	// default to 'q' for a new created view.
	QuitRunes []rune

	// QuitKeys are the special-key-inputs which trigger a quit event.
	// They default to ctrl-c and ctrl-d for a new created view.
	QuitKeys []tcell.Key
}

func (tt *Triggers) resolveRuneEventType(r rune) evtType {
	for _, _r := range tt.QuitRunes {
		if _r == r {
			return quit
		}
	}
	return none
}

type defaultFactory struct{}

func (f *defaultFactory) NewScreen() (tcell.Screen, error) {
	return tcell.NewScreen()
}

func (f *defaultFactory) NewSimulationScreen(
	s string,
) tcell.SimulationScreen {
	return tcell.NewSimulationScreen(s)
}

type screenFactoryer interface {
	NewScreen() (tcell.Screen, error)
	NewSimulationScreen(string) tcell.SimulationScreen
}

var screenFactory screenFactoryer = &defaultFactory{}

func toInt(kk []tcell.Key) []int {
	ii := make([]int, len(kk))
	for i, k := range kk {
		ii[i] = int(k)
	}
	return ii
}
