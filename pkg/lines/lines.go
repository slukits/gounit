// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package lines provides a simple terminal-UI where the UI is
// interpreted as an ordered set of lines.  lines hides event-handling
// and screen-synchronization from its user.  Making the components of
// a view concurrency save in terms that it can be manipulated while an
// event is processed at the same time would add significant overhead
// and complexity to the view.  To avoid this overhead and race
// conditions this package was designed as follows.
//
// reg := lines.New()
//
// will return a so called "listener register" which may be used to
// register for events:
//
// reg.Resize(func(v *lines.View) { v.Lines.Get(0).Set("line 0") })
//
// the above line will effectively print "line 0" into the first line of
// a terminal once the initial resize-event was emitted after a call of
//
// reg.Listen()
//
// the later starts the event loop and blocks until a Quit-event was
// received or reg.QuitListening() was called.
//
// reg.Update(func(v *lines.View) { v.Lines.Get(0).Set("updated 0") })
//
// the Update method posts an update event into the event-loop and calls
// given listener back once it is polled.  In order to react on user
// input listeners may be registered for runes or special keys as they
// are recognized and provided by the underlying *tcell* package
//
// func help(v *lines.View, m tcell.ModMask) {
//     v.Lines.Get(0).Set("some help-text in first line")
// }
// reg.Key(help, tcell.KeyF1)
// reg.Rune(help, 'h')
//
// i.e. help is called back if the user presses either the F1 or the H
// key.
//
// reg.Runes(func (v *lines.View, r rune) {
//     v.Lines.Get(0).Set("received rune-input: "+string(r))
// })
//
// Runes suppresses all registered Rune-events and provides received
// rune input to registered Runes-listener until a non-rune is received.
//
// reg.Quit(func() { fmt.Println("good by") })
//
// a Quit-listener is called iff a quit event is received which happens
// if 'q', ctrl-c or ctrl-d is received.
//
// NOTE to avoid races a view must be only manipulated within an
// event-listener and if the event-listener returns no further
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
	return &Register{
		view:   view,
		keys:   &sync.Mutex{},
		kk:     map[tcell.Key]func(*View, tcell.ModMask){},
		runes:  &sync.Mutex{},
		rr:     map[rune]func(*View){},
		other:  &sync.Mutex{},
		Synced: make(chan bool, 1),
	}, nil
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
	return &Register{
		view:   view,
		keys:   &sync.Mutex{},
		kk:     map[tcell.Key]func(*View, tcell.ModMask){},
		runes:  &sync.Mutex{},
		rr:     map[rune]func(*View){},
		other:  &sync.Mutex{},
		Synced: make(chan bool, 1),
	}, view.lib.(tcell.SimulationScreen), nil
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
