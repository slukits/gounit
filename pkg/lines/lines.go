// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package lines provides a well tested, robust against race conditions,
// simple, easy to use terminal-UI where the UI is interpreted as an
// ordered set of lines.  *lines* hides event-polling and
// screen-synchronization from its user.  Making the components of a
// view concurrency save in the sense that it can be manipulated by
// several event-listeners while screen synchronization is processed at
// the same time would add significant overhead and complexity to the
// view.  To avoid this overhead and be still robust against race
// conditions this package was designed not around a view/screen/window
// but around event-handling:
//
// reg := lines.New()
//
// will return a so called "listener register" which may be used to
// register call-back functions for events:
//
// reg.Resize(func(v *lines.View) { v.LL().Get(0).Set("line 0") })
//
// The above line will effectively print "line 0" into the first line of
// a terminal once the initial resize-event was emitted after a call of
//
// reg.Listen()
//
// The later starts the event loop and blocks until a Quit-event was
// received or reg.QuitListening() was called.
//
// reg.Update(func(v *lines.View) { v.LL().Get(0).Set("updated 0") })
//
// The Update method posts an update event into the event-loop and calls
// given listener back once it is polled.  I.e. Update provides a
// programmatically way to update the screen without user triggered
// events. To react on user input listeners may be registered for runes
// or special keys as they are recognized and provided by the underlying
// *tcell* package
//
// func help(v *lines.View, m tcell.ModMask) {
//     v.LL().Get(0).Set("some help-text in first line")
// }
// reg.Key(help, tcell.KeyF1)
// reg.Rune(help, 'h')
//
// i.e. help is called back if the user presses either the F1 or the H
// key.
//
// reg.KeyBoard(func (v *lines.View, r rune, k tcell.Key) (stop bool) {
//     if r != rune(0) {
//         v.LL().Get(0).Set("received rune-input: "+string(r))
//     }
//     switch k {
//     case tcell.KeyESC, tcell.KeyEnter:
//         return true
//     default:
//         return false
//     }
// })
//
// KeyBoard suppresses all registered Rune- and Key-events (except for
// the quit event) and provides received rune/key input to registered
// Runes-listener until it returns true.
//
// reg.Quit(func() { fmt.Println("good by") })
//
// a Quit-listener is called iff a quit event is received which happens
// by default if 'q', ctrl-c or ctrl-d is received.
//
// NOTE to avoid race conditions a view must be only manipulated within
// an event-listener and if the event-listener returns no further
// manipulations must happen.  Simply never (!) keep a view instance
// around.   If concurrent view-manipulations are done inside a listener
// the listener is responsible that they are performed in a concurrency
// save way.
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
		view:   view,
		kk:     map[tcell.Key]func(*View, tcell.ModMask){},
		rr:     map[rune]func(*View){},
		mutex:  &sync.Mutex{},
		Synced: make(chan bool, 1),
		Keys:   DefaultKeys,
	}
	reg.Keys = DefaultKeys.copy(&reg)
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
	reg.Keys = DefaultKeys.copy(&reg)
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
