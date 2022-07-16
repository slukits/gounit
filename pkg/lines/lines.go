// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package lines provides an unopinionated, well tested and documented,
// robust against race conditions, simple, easy to use terminal-UI.  The
// terminal is interpreted as an ordered set of lines which you in the
// future might even can split into columns and rows.  Its
// implementation is motivated by my experience with other small
// ui-libraries which try to make it convenient to implement an ui.
// Often enough these libraries impose unwanted behavior to its user
// which is difficult to bypass, i.e. the convenience gets lost.  And
// also often enough I encountered "strange behavior", i.e. it seems
// kinda difficult to get right.  Especially when it comes to concurrent
// usage.
//
// *lines* only imposes three things to its user which you might want to
// consider before you decide for it.
//
// Firstly the keys ctrl-c and ctrl-d quit the application. Always.
//
// Secondly *lines* wraps the package tcell which does the heavy lifting
// on the terminal side.  I didn't make the effort wrap the constants
// and types which are defined by tcell and are used for event-handling
// and styling.  I.e. you will have to make yourself acquainted with
// tcell's *Key* constants its *ModeMap* constants, its *AttrMask*
// constants, its *Style* type and color handling as needed.
//
// Thirdly an architectural choice.  A typical ui-library has generally
// two functions:
//
// - providing user input events
//
// - a screen/display/window/view one can print/draw to.
//
// In lines the View type represents the instance where you write your
// output to hence I will use further on the term view.  One of go's
// killer features is concurrency.  Using a view concurrently is either
// prone to rase conditions or adds considerable complexity and overhead
// to a view's implementation if it were to be concurrency save.  To
// avoid both I decided to design *lines* around the event-handling and
// not around the view which seems to be more common.  I.e.
//
//     ee := lines.New()
//
// will return an *Events* instance which may be used to register
// call-back functions for events:
//
//     ee.Resize(func(e *lines.Env) { e.LL().Get(0).Set("line 0") })
//
// The above line will effectively print "line 0" into the first line of
// a terminal once the initial resize-event was emitted after a call of
//
//     ee.Listen()
//
// The later starts the event loop and blocks until a Quit-event was
// received or ee.QuitListening() was called.  Note a lines.Env (short
// for environment) instance is provided with every event-listener
// callback.  Env embeds the View-type, i.e. has all methods of the View
// plus some aspects for the communication between the event-lister and
// the reporting Events-instance, i.e. e.StopBubbling().
//
//     ee.Update(func(e *lines.Env) {
//         e.LL().Get(0).Set("updated 0")
//     })
//
// The Update method posts an update event into the event-loop and calls
// given listener back once it is polled.  I.e. Update provides a
// programmatically way to update the view without user triggered
// events. To react on user input listeners may be registered for runes
// or special keys as they are recognized and provided by the underlying
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
// Keyboard suppresses all registered Rune- and Key-events (except -
// remember :) - for the quit event) and provides received rune/key
// input to registered Keyboard-listener until it is removed.
//
//     ee.Quit(func() { fmt.Println("good by") })
//
// A Quit-listener is called if a quit event is received which happens
// by default if 'q', ctrl-c or ctrl-d is received.  Note you can remove
// the 'q'-rune from the quit-event handling (but not ctrl-c or ctrl-d).
//
// The underlying *tcell* library's event-loop already provides an
// serialization mechanism which is leveraged to make this package
// robust against race conditions.  If you can resist the temptation to
// let an Env-instance leave a listener's implementation and make sure
// that all manipulations of a view are finished when the event-listener
// returns then a view is concurrency save by design.  If you want in
// response to an event manipulate a view from more than one go-routine
// *you* must take care that it is done in a concurrency save way which
// is very difficult if you have not studied the implementation of the
// view.  If you have cpu/io-heavy operations whose result should go to
// the screen then send them of in their own go routine (without the
// Env-instance) which at the end registers an update event which once
// called back prints its findings to the view.  Note registering
// event-listeners, i.e. the Events-type, *is* implemented concurrency
// save! E.g.:
//
//     // can not be run in the go-playground since ee.Listen() is blocking
//
//     import "github.com/slukits/lines"
//
//     func countTextFilesOnMyComputer() {
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
//         // NOTE the Env-instance does not leave the listener!
//         go countTextFilesOnMyComputer()
//         e.Statusbar().Set("counting text files").Busy()
//     }
//
//     func main() {
//         ee := lines.New()
//         ee.Key(tcell.KeyF5, 0, countTextFilesListener)
//         ee.Listen()
//     }
//
// assuming an appropriate actual implementation this program will print
// the text-files count to the statusbar in a non-blocking
// race-condition free manner once a user presses F5.
//
// If you can live with these three aspects everything else is optional
// and at your service as needed. E.g. as long as you don't call
// Statusbar on a view, a view doesn't have a statusbar.  If you call
// Statusbar you get one with useable defaults.  If you don't like these
// defaults you can change them ...
//
// Features
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
		view:     view,
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
		view:   view,
		ll:     NewListeners(DefaultFeatures),
		mutex:  &sync.Mutex{},
		Synced: make(chan bool, 1),
	}
	ee.Features = DefaultFeatures.Copy()
	return &ee, view.lib.(tcell.SimulationScreen), nil
}

// newView returns a new View instance or nil and an error in case
// tcell's screen-creation or its initialization fails.
func newView() (*View, error) {
	lib, err := screenFactory.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := lib.Init(); err != nil {
		return nil, err
	}
	v := &View{lib: lib, Synced: make(chan bool, 1)}
	v.ll = &Lines{vw: v}
	return v, nil
}

// newSim returns a new View instance wrapping tcell's simulation
// screen for testing purposes.
func newSim() (*View, error) {
	lib := screenFactory.NewSimulationScreen("")
	if err := lib.Init(); err != nil {
		return nil, err
	}
	v := &View{lib: lib, Synced: make(chan bool, 1)}
	v.ll = &Lines{vw: v}
	return v, nil
}
