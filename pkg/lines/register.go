// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Register allows to register for user-input events and posting
// update-events.  Registering of events is concurrency save.
type Register struct {
	view      *View
	mutex     *sync.Mutex
	ll        *Listeners
	resize    func(*View)
	quit      func()
	isPolling bool

	// Synced sends a message after a the screen synchronization
	// following a reported event.
	Synced chan bool

	// Ev provides the currently reported tcell-event.  NOTE this value
	// is lost the moment an event-listener called for this event
	// returns.
	Ev tcell.Event

	// Features are the keys and runes which are used for "internal"
	// event handling, e.g. the keys/runes for the quit event are q,
	// ctrl-c and ctrl-d.  The Features instance allows you to add and
	// remove features and manipulate their event-keys in a consistent
	// way.  Features default to *DefaultFeatures*.
	Features *Features
}

// IsPolling returns true if listener register is polling in the event
// loop.
func (rg *Register) IsPolling() bool {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	return rg.isPolling
}

// Listen blocks and starts polling from the event loop reporting
// received events to registered listeners.  Listen returns if either a
// quit-event was received ('q', ctrl-c, ctrl-d input) or QuitListening
// was called.
func (rg *Register) Listen() {
	rg.startPolling()
	for {
		ev := rg.view.lib.PollEvent()

		select {
		case <-rg.Synced:
		default:
		}

		switch ev := ev.(type) {
		case nil: // event-loop ended
			return
		case *quitEvent:
			rg.stopPolling()
			if rg.quit != nil {
				rg.Ev = ev
				rg.quit()
			}
			rg.quitListening()
		case *tcell.EventResize:
			if rg.view.resize() {
				// TODO: make report cancelable after a set timeout.
				rg.report(ev)
			}
			rg.view.ensureSynced(false)
			rg.Synced <- true
		default:
			quit := rg.report(ev)
			if quit {
				rg.stopPolling()
				rg.quitListening()
				return
			}
			rg.view.ensureSynced(true)
			rg.Synced <- true
		}
	}
}

func (rg *Register) startPolling() {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	rg.isPolling = true
}

func (rg *Register) stopPolling() {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	rg.isPolling = false
}

// Resize registers given listener for the resize event.  Note starting
// the event-loop by calling *Listen* will trigger a mandatory initial
// resize event.
func (rg *Register) Resize(listener func(*View)) {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	rg.resize = listener
}

// Quit registers given listener for the quit event which is triggered
// by 'r'-rune, ctrl-c and ctrl-d.
func (rg *Register) Quit(listener func()) {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	rg.quit = listener
}

// Update posts a new event into the event loop which calls once it is
// its turn given listener.  Update fails if the event-loop is full
// returned error will wrap tcell's *PostEven* error.  Update is an
// no-op if listener is nil.
func (rg *Register) Update(listener func(*View)) error {
	if listener == nil {
		return nil
	}
	evt := &updateEvent{
		when:     time.Now(),
		listener: listener,
	}
	if err := rg.view.lib.PostEvent(evt); err != nil {
		return fmt.Errorf(ErrUpdateFmt, err)
	}
	return nil
}

// ErrUpdateFmt is the error message for a failing update-event post.
var ErrUpdateFmt = "can't post event: %w"

type updateEvent struct {
	when     time.Time
	listener func(*View)
}

func (u *updateEvent) When() time.Time { return u.when }

func (rg *Register) Rune(r rune, l Listener) error {
	return rg.ll.Rune(r, l)
}

// Keyboard listener shadows all other rune/key listeners until it is
// removed by Keyboard(nil).
func (rg *Register) Keyboard(l KBListener) {
	rg.ll.Keyboard(l)
}

func (rg *Register) Key(k tcell.Key, m tcell.ModMask, l Listener) error {
	return rg.ll.Key(k, m, l)
}

// QuitListening posts a quit event ending the event-loop, i.e.
// IsPolling will be false.
func (rg *Register) QuitListening() {
	if rg.isPolling {
		rg.view.lib.PostEvent(&quitEvent{when: time.Now()})
		return
	}
	rg.quitListening()
}

func (rg *Register) quitListening() {
	rg.view.lib.Fini()
	close(rg.Synced)
}

type quitEvent struct {
	when time.Time
}

func (u *quitEvent) When() time.Time { return u.when }

func (rg *Register) report(ev tcell.Event) (quit bool) {
	rg.Ev = ev
	if rg.view.ToSmall() {
		if ev, ok := ev.(*tcell.EventKey); ok {
			if rg.isQuitEvent(ev) {
				if listener := rg.quitListener(ev); listener != nil {
					listener()
				}
				return true
			}
		}
		return false
	}
	switch ev := ev.(type) {
	case *tcell.EventResize:
		if listener := rg.resizeListener(); listener != nil {
			listener(rg.view)
		}
	case *tcell.EventKey:
		if rg.isQuitEvent(ev) {
			if listener := rg.quitListener(ev); listener != nil {
				listener()
			}
			return true
		}
		if kbl := rg.ll.KBListener(); kbl != nil {
			kbl(rg.view, ev.Rune(), ev.Key(), ev.Modifiers())
			return
		}
		if l, ok := rg.ll.RuneListenerOf(ev.Rune()); ok {
			l(rg.view)
		}
		if l, ok := rg.ll.KeyListenerOf(ev.Key(), ev.Modifiers()); ok {
			l(rg.view)
		}
	case *updateEvent:
		ev.listener(rg.view)
	}
	return false
}

func (rg *Register) resizeListener() func(*View) {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	return rg.resize
}

func (rg *Register) isQuitEvent(ev *tcell.EventKey) bool {
	if ev.Key() == tcell.KeyRune && ev.Rune() != 'q' {
		return false
	}
	if ev.Key() != tcell.KeyRune && ev.Key() != tcell.KeyCtrlC &&
		ev.Key() != tcell.KeyCtrlD {
		return false
	}
	return true
}

func (rg *Register) quitListener(ev *tcell.EventKey) func() {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	return rg.quit
}
