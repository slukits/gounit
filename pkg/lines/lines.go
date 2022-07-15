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
// two functions: providing user input events and a display one can
// print/draw to.  The later is in *lines* represented by the type
// *View* thus I will use the term "view" to refer to a terminal screen
// or emulator.  One of go's killer features is concurrency.  Using an
// view concurrently is either prone to rase conditions or adds
// considerable complexity and overhead to a view's implementation if it
// were to be concurrency save.  To avoid both I decided to design
// *lines* around the event-handling and not around the view which seems
// to be more common.  I.e.
//
//     reg := lines.New()
//
// will return a so called "listener register" which may be used to
// register call-back functions for events:
//
//     reg.Resize(func(v *lines.View) {
//         v.LL().Get(0).Set("line 0")
//     })
//
// The above line will effectively print "line 0" into the first line of
// a terminal once the initial resize-event was emitted after a call of
//
//     reg.Listen()
//
// The later starts the event loop and blocks until a Quit-event was
// received or reg.QuitListening() was called.
//
//     reg.Update(func(v *lines.View) {
//         v.LL().Get(0).Set("updated 0")
//     })
//
// The Update method posts an update event into the event-loop and calls
// given listener back once it is polled.  I.e. Update provides a
// programmatically way to update the view without user triggered
// events. To react on user input listeners may be registered for runes
// or special keys as they are recognized and provided by the underlying
// *tcell* package
//
//     func help(v *lines.View) {
//         v.Statusbar().Set("some help-text in the statusbar")
//     }
//     reg.Key(tcell.KeyF1, 0, help)
//     reg.Rune('H', help)
//
// I.e. *help* is called back if the user presses either the F1 or the H
// key.
//
//     reg.Keyboard(func (v *lines.View, r rune, k tcell.Key, m tcell.ModMask) {
//         if r != rune(0) {
//             v.Message().Styledf(Centered, "received rune-input: %c", r)
//             return
//         }
//         reg.Keyboard(nil)
//     })
//
// KeyBoard suppresses all registered Rune- and Key-events (except -
// remember :) - for the quit event) and provides received rune/key
// input to registered Keyboard-listener until it is removed.
//
//     reg.Quit(func() { fmt.Println("good by") })
//
// A Quit-listener is called if a quit event is received which happens
// by default if 'q', ctrl-c or ctrl-d is received.  Note you can remove
// the 'q'-rune from the quit-event handling.
//
// The underlying *tcell* library's event-loop already provides an
// serialization mechanism which is leveraged to make this package
// robust against race conditions.  If you can resist the temptation to
// keep a view around outside an event-handler and make sure that all
// manipulations of a view are finished when the event-listener returns
// then a View is concurrency save by design.  If you want in response
// to an event manipulate a view from more than one go-routine *you*
// must take care that it is done in a concurrency save way which is
// very difficult if you have not studied the implementation of the
// view.  If you have cpu/io-heavy operations whose result should go to
// the screen then send them of in their own go routine which at the end
// registers an update event which once called back prints its findings
// to the view.  Note registering event-listeners, i.e. the
// Register-type, *is* implemented concurrency save! E.g.:
//
//     // can not be run in the go-playground since reg.Listen() is blocking
//
//     import "github.com/slukits/lines"
//
//     func countTextFilesOnMyComputer() {
//         n := func() int {
//              // actual implementation which defaults for the sake
//              // of an executable example to
//              return 42
//         }()
//         reg.Update(func(v *View) {
//             v.Statusbar().Setf("found %d files", n)
//         })
//     }
//
//     func countTextFilesListener(v *View) {
//         // NOTE the view does not leave the listener!
//         go countTextFilesOnMyComputer()
//         v.Statusbar().Set("counting text files").Busy()
//     }
//
//     func main() {
//         reg := lines.New()
//         reg.Key(tcell.KeyF5, 0, countTextFilesListener)
//         reg.Listen()
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
func New() (*Register, error) {
	view, err := newView()
	if err != nil {
		return nil, err
	}
	reg := Register{
		view:     view,
		kk:       map[tcell.Key]func(*View, tcell.ModMask){},
		rr:       map[rune]func(*View){},
		mutex:    &sync.Mutex{},
		Synced:   make(chan bool, 1),
		Features: DefaultFeatures,
	}
	reg.Features = DefaultFeatures.Copy()
	return &reg, nil
}

// Sim returns a listener register providing a view with tcell's
// simulation screen.  Since the wrapped tcell screen is private it is
// returned as well to facilitate desired mock-ups.  Sim fails iff
// tcell's screen-initialization fails.
func Sim() (*Register, tcell.SimulationScreen, error) {
	view, err := newSim()
	if err != nil {
		return nil, nil, err
	}
	reg := Register{
		view:   view,
		kk:     map[tcell.Key]func(*View, tcell.ModMask){},
		rr:     map[rune]func(*View){},
		mutex:  &sync.Mutex{},
		Synced: make(chan bool, 1),
	}
	reg.Features = DefaultFeatures.Copy()
	return &reg, view.lib.(tcell.SimulationScreen), nil
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
