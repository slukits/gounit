// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package lines provides an unopinionated, well tested and documented,
// robust against race conditions, simple, easy to use terminal-UI.  The
// terminal is interpreted as an ordered set of lines which you in the
// future might even can split into columns and rows.  Its
// implementation is motivated by my experience with other small
// ui-libraries which try to make it convenient to implement quickly an
// ui.
//
// lines only imposes two things to its user which you might want to
// consider before you decide for it.
//
// Firstly the keys ctrl-c and ctrl-d quit the application. Always.
//
// Secondly lines wraps the package tcell which does the heavy lifting
// on the terminal side.  I didn't make the effort to wrap the constants
// and types which are defined by tcell and are used for event-handling
// and styling.  I.e. you will have to make yourself acquainted with
// tcell's Key constants its ModeMap constants, its AttrMask constants,
// its Style type and Color handling as needed.
//
// Everything else is at your service if you request it otherwise its
// not in you way.  For example if you don't ask for a message-bar you
// dont have one.  If you ask for a message bar you get one with
// reasonable defaults.  If you don't like these defaults you can change
// them...
//
// Events
//
// A typical ui-library has generally two functions:
//
// - providing user input events
//
// - a screen/display/window/view one can print/draw to.
//
// In lines the terminal-screen is accessed through a provided
// environment instance to event-listeners.  One of go's killer features
// is concurrency.  Using a screen concurrently is either prone to rase
// conditions or adds considerable complexity and overhead to a Screen's
// implementation if it were to be concurrency save.  To avoid both I
// decided to design lines around event-handling and not around the
// screen which seems to be more common.  I.e.
//
//     import "github.com/slukits/lines"
//
//     ee := lines.New()
//
// will return an Events instance which may be used to register
// call-back functions for events:
//
//     ee.Resize(func(e *lines.Env) { e.LL().Get(0).Set("line 0") })
//
// The above line will effectively print "line 0" into the first line of
// a terminal once the initial resize-event was emitted (and with every
// further resize event if not changed) after a call of
//
//     ee.Listen()
//
// The later starts the event loop and blocks until a Quit-event was
// received or ee.QuitListening() was called.  An Env-instance
// (environment) encapsulates a Screen instance and provides a Screen's
// public API (not the screen itself).  Env also provides information
// about the current event and means to communicate back to the
// reporting Events-instance.  See the documentation of the Screen-Type
// to learn what you can do with a Screen.  If you use an Env instance
// in an other go routine you will get most like a nil pointer panic.
// If you want concurrency use:
//
//     ee.Update(func(e *lines.Env) {
//         e.LL().Get(0).Set("updated 0")
//     })
//
// The Update method posts an update event into the event-loop which
// calls given listener back once it is polled.  I.e. Update provides a
// programmatically way to trigger an environment providing callback
// without user input.  With this feature we can send cpu/io-heavy
// operation of in their own go-routine and this go-routine once done
// registers an update event to inform the user about its findings:
//
//     // can not be run in the go-playground since ee.Listen() is blocking
//
//     import "github.com/slukits/lines"
//
//     func countTextFilesOnMyComputer(ee *Events) {
//         n := func() int {
//              // actual implementation which defaults for the sake
//              // of an executable example to
//              return 42
//         }()
//         ee.Update(func(e *lines.Env) {
//             e.Statusbar().Setf("found %d files", n)
//         })
//     }
//
//     func countTextFilesListener(e *lines.Env) {
//         // NOTE the Env-instance is not passed on to the go routine!
//         // But a property provided by e you can pass on.
//         go countTextFilesOnMyComputer(e.EE)
//         e.Statusbar().Set("counting text files").Busy()
//     }
//
//     func main() {
//         ee := lines.New()
//         ee.Key(tcell.KeyF5, 0, countTextFilesListener)
//         ee.Listen()
//     }
//
// The rule of thump is here: environment properties you can safely pass
// on to an other go routine; return values of environment methods you
// can't if you want to avoid race conditions.
//
// To react on user input listeners may be registered for runes or
// special keys as they are recognized and provided by the underlying
// *tcell* package
//
//     func help(e *lines.Env) {
//         e.Statusbar().Set("some help-text in the statusbar")
//     }
//     ee.Key(tcell.KeyF1, 0, help)
//     ee.Rune('H', help)
//
// I.e. *help* is called back if the user presses either the F1 or the H
// key.
//
//     ee.Keyboard(func (e *lines.Env, r rune, k tcell.Key, m tcell.ModMask) {
//         if r != rune(0) {
//             e.Message().Styledf(
//                 Centered, "received rune-input: %c", r)
//             return
//         }
//         ee.Keyboard(nil)
//     })
//
// Keyboard suppresses all registered rune- and key-events (except for
// the quit event) and provides received rune/key input to registered
// Keyboard-listener until it is removed.
//
//     ee.Quit(func() { fmt.Println("good by") })
//
// A Quit-listener is called if a quit event is received which happens
// by default if 'q', ctrl-c or ctrl-d is received.  Note you can remove
// the 'q'-rune from the quit-event handling (but not ctrl-c or ctrl-d).
//
// Listeners
//
// Events encapsulates Listeners where we have registered listeners for
// "global" events in the previous section.  With
//
//     ll := NewListeners(nil)
//
// you can create your own Listeners-instance and register for events as
// we did before.  To make use of ll we can assign this set of
// event-listener registrations to an Screen Component.  Most things
// which are returned by environment methods are Screen Components, e.g.
//
//    e.MessageBar().Listeners = ll
//
// Now we have set our listeners to the Screen Component message bar
// which still only waists memory because our message bar can't receive
// the focus.  To make the above actually do something we need last but
// not least make use of Features.
//
// Features
//
// In order to use features like focusing or scrolling we need to turn
// these features on.  lines could try to be smart and reason "if you
// want to receive key-events on the message bar the screen must have
// the feature focusing turned on".  Since turning focusing on changes
// the activated key-bindings as well as the layout and behavior of your
// application --- none of which you have asked for --- you need to ask
// for it
//
//     e.Features.Key(
//         FtFocusNext, tcell.KeyTab, tcell.ModNone)
//     e.Features.Key(
//         FtFocusNext, tcell.KeyTab, tcell.ModShift)
//
// Now the environment's Screen has "focusing" turned on and the message
// bar can receive the focus which activates its event listeners.  There
// are predefined feature sets with common defaults to keep things easy
// for you.  E.g.
//
//     e.LL().Features = NewFeatures(Focusing, Scrolling)
//
// will bind the page up/down keys to scroll up and down and the tab-key
// like above to focus lines of the currently focused Screen Component
// providing (screen) lines.  See the documentation of the Features-type
// to learn how to change the defaults of features sets.
package lines

import (
	"sync"

	"github.com/gdamore/tcell/v2"
)

// New returns a listener register providing a view to its event
// listeners or nil and an error in case tcell's screen-creation or its
// initialization fails.
func New() (*Events, error) {
	view, err := newView()
	if err != nil {
		return nil, err
	}
	ee := Events{
		scr:      view,
		ll:       NewListeners(DefaultFeatures),
		mutex:    &sync.Mutex{},
		Synced:   make(chan bool, 1),
		Features: DefaultFeatures,
	}
	ee.Features = DefaultFeatures.Copy()
	return &ee, nil
}

// Sim returns a listener register providing a view with tcell's
// simulation screen.  Since the wrapped tcell screen is private it is
// returned as well to facilitate desired mock-ups.  Sim fails iff
// tcell's screen-initialization fails.
func Sim() (*Events, tcell.SimulationScreen, error) {
	view, err := newSim()
	if err != nil {
		return nil, nil, err
	}
	ee := Events{
		scr:    view,
		ll:     NewListeners(DefaultFeatures),
		mutex:  &sync.Mutex{},
		Synced: make(chan bool, 1),
	}
	ee.Features = DefaultFeatures.Copy()
	return &ee, view.lib.(tcell.SimulationScreen), nil
}

// newView returns a new Screen instance or nil and an error in case
// tcell's screen-creation or its initialization fails.
func newView() (*Screen, error) {
	lib, err := screenFactory.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := lib.Init(); err != nil {
		return nil, err
	}
	v := &Screen{lib: lib}
	v.ll = &Lines{scr: v}
	return v, nil
}

// newSim returns a new Screen instance wrapping tcell's simulation
// screen for testing purposes.
func newSim() (*Screen, error) {
	lib := screenFactory.NewSimulationScreen("")
	if err := lib.Init(); err != nil {
		return nil, err
	}
	v := &Screen{lib: lib}
	v.ll = &Lines{scr: v}
	return v, nil
}
