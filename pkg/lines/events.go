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

// Events allows to listen for user-input event which are then reported
// to registered listeners.  It also manages behind the scenes the
// screen synchronization.
type Events struct {
	scr         *Screen
	mutex       *sync.Mutex
	ll          *Listeners
	resize      Listener
	quit        func(e *Env)
	isListening bool
	reported    func()
	t           *Testing

	// Synced sends a message after a the screen synchronization
	// following a reported event.
	Synced chan bool

	// Features are the keys and runes which are used for "internal"
	// event handling, e.g. the keys/runes for the quit event are q,
	// ctrl-c and ctrl-d.  The Features instance allows you to add and
	// remove features and manipulate their event-keys in a consistent
	// way.  Features default to *DefaultFeatures*.
	Features *Features
}

// IsListening returns true if given Events polling from the event loop.
func (rg *Events) IsListening() bool {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	return rg.isListening
}

// Listen blocks and starts polling from the event loop reporting
// received events to registered listeners.  Listen returns if either a
// quit-event was received ('q', ctrl-c, ctrl-d input) or QuitListening
// was called.  NOTE in testing Listen is non-blocking, i.e. returns
// after the initial resize was processed.
func (ee *Events) Listen() {
	if ee.t != nil {
		ee.t.listen()
		return
	}
	ee.listen()
}

func (ee *Events) listen() {
	if !ee.startPolling() { // ignore subsequent calls of Listen
		return
	}
	for {
		ev := ee.scr.lib.PollEvent()

		select {
		case <-ee.Synced:
		default:
		}

		switch ev := ev.(type) {
		case nil: // event-loop ended
			return
		case *quitEvent:
			ee.stopPolling()
			if ee.quit != nil {
				ee.quit(ee.env(ev))
			}
			ee.quitListening()
		case *tcell.EventResize:
			if ee.scr.resize() {
				// TODO: make report cancelable after a set timeout.
				ee.report(ev)
			}
			ee.scr.ensureSynced(false)
			ee.Synced <- true
		default:
			if quit := ee.report(ev); quit {
				ee.stopPolling()
				ee.quitListening()
				return
			}
			ee.scr.ensureSynced(true)
			ee.Synced <- true
		}
	}
}

func (ee *Events) startPolling() bool {
	ee.mutex.Lock()
	defer ee.mutex.Unlock()
	if ee.isListening {
		return false
	}
	ee.isListening = true
	return true
}

func (ee *Events) stopPolling() {
	ee.mutex.Lock()
	defer ee.mutex.Unlock()
	ee.isListening = false
}

// Reported calls back if an event was reported for logging and testing.
func (ee *Events) Reported(listener func()) {
	ee.reported = listener
}

// Resize registers given listener for the resize event.  Note starting
// the event-loop by calling *Listen* will trigger a mandatory initial
// resize event.
func (ee *Events) Resize(l Listener) {
	ee.mutex.Lock()
	defer ee.mutex.Unlock()
	ee.resize = l
}

// Quit registers given listener for the quit event which is triggered
// by 'r'-rune, ctrl-c and ctrl-d.
func (ee *Events) Quit(listener func(*Env)) {
	ee.mutex.Lock()
	defer ee.mutex.Unlock()
	ee.quit = listener
}

// Update posts a new event into the event loop which calls once it is
// its turn given listener.  Update fails if the event-loop is full
// returned error will wrap tcell's *PostEvent* error.  Update is an
// no-op if listener is nil.  NOTE in testing Update returns after the
// event was processed.
func (ee *Events) Update(l Listener) error {
	if l == nil {
		return nil
	}
	if ee.t != nil && !ee.isListening {
		ee.t.listen()
	}
	evt := &updateEvent{
		when:     time.Now(),
		listener: l,
	}
	if err := ee.scr.lib.PostEvent(evt); err != nil {
		return fmt.Errorf(ErrUpdateFmt, err)
	}
	if ee.t != nil {
		ee.t.waitForSynced("test: update: sync timed out")
		ee.t.checkTermination()
	}
	return nil
}

// ErrUpdateFmt is the error message for a failing update-event post.
var ErrUpdateFmt = "can't post event: %w"

type updateEvent struct {
	when     time.Time
	listener Listener
}

func (u *updateEvent) When() time.Time { return u.when }

// Rune registers a given listener for given rune-event.  It fails if
// already a listener is registered for given rune-event.
func (ee *Events) Rune(r rune, l Listener) error {
	return ee.ll.Rune(r, l)
}

// Keyboard listener shadows all other rune/key listeners until it is
// removed by Keyboard(nil).
func (ee *Events) Keyboard(l KBListener) {
	ee.ll.Keyboard(l)
}

// Key registers given listener for given key/mode-event.  It fails if
// already a listener is registered for given key/mode combination.
func (ee *Events) Key(k tcell.Key, m tcell.ModMask, l Listener) error {
	return ee.ll.Key(k, m, l)
}

// QuitListening posts a quit event ending the event-loop, i.e.
// IsPolling will be false.
func (ee *Events) QuitListening() {
	if ee.isListening {
		ee.scr.lib.PostEvent(&quitEvent{when: time.Now()})
		if ee.t != nil {
			ee.t.waitForSynced("test: quit listening: sync timed out")
		}
		return
	}
	ee.quitListening()
}

func (ee *Events) quitListening() {
	if ee.t != nil {
		ee.t.beforeFinalize()
	}
	ee.scr.lib.Fini()
	close(ee.Synced)
}

type quitEvent struct {
	when time.Time
}

func (u *quitEvent) When() time.Time { return u.when }

func (ee *Events) report(ev tcell.Event) (quit bool) {
	if ee.scr.ToSmall() {
		return ee.reportToSmall(ev)
	}
	switch ev := ev.(type) {
	case *tcell.EventResize:
		if listener := ee.resizeListener(); listener != nil {
			env := ee.env(ev)
			listener(env)
			ee.reportReported(env)
		}
	case *tcell.EventKey:
		return ee.reportKeyEvent(ev)
	case *updateEvent:
		env := ee.env(ev)
		ev.listener(env)
		ee.reportReported(env)
	}
	return false
}

// reportToSmall handles reporting an event in case the view is to
// small, i.e. only reports the quit-event.
func (ee *Events) reportToSmall(ev tcell.Event) bool {
	if ev, ok := ev.(*tcell.EventKey); ok {
		if ee.isQuitEvent(ev) {
			if listener := ee.quitListener(ev); listener != nil {
				env := ee.env(ev)
				listener(env)
				ee.reportReported(env)
			}
			return true
		}
	}
	return false
}

func (ee *Events) reportKeyEvent(ev *tcell.EventKey) bool {
	if ee.isQuitEvent(ev) {
		if listener := ee.quitListener(ev); listener != nil {
			env := ee.env(ev)
			listener(env)
			ee.reportReported(env)
		}
		return true
	}
	if kbl := ee.ll.KBListener(); kbl != nil {
		env := ee.env(ev)
		kbl(env, ev.Rune(), ev.Key(), ev.Modifiers())
		ee.reportReported(env)
		return false
	}
	if l, ok := ee.ll.RuneListenerOf(ev.Rune()); ok {
		env := ee.env(ev)
		l(env)
		ee.reportReported(env)
	}
	if l, ok := ee.ll.KeyListenerOf(ev.Key(), ev.Modifiers()); ok {
		env := ee.env(ev)
		l(env)
		ee.reportReported(env)
	}
	return false
}

func (ee *Events) env(ev tcell.Event) *Env {
	return &Env{
		scr: ee.scr,
		EE:  ee,
		Evn: ev,
	}
}

// reportReported is for testing purposes reporting back each time an
// event was reported allowing the Events-fixture-implementation to
// count reported events and end the event-loop automatically after a
// certain amount of reported events.
func (ee *Events) reportReported(env *Env) {
	if env != nil {
		env.reset()
	}
	if l := ee.reportedListener(); l != nil {
		l()
	}
}

func (ee *Events) reportedListener() func() {
	ee.mutex.Lock()
	defer ee.mutex.Unlock()
	return ee.reported
}

func (ee *Events) resizeListener() Listener {
	ee.mutex.Lock()
	defer ee.mutex.Unlock()
	return ee.resize
}

func (ee *Events) isQuitEvent(ev *tcell.EventKey) bool {
	// TODO: use Features to determine the quit event.
	if ev.Key() == tcell.KeyRune && ev.Rune() != 'q' {
		return false
	}
	if ev.Key() != tcell.KeyRune && ev.Key() != tcell.KeyCtrlC &&
		ev.Key() != tcell.KeyCtrlD {
		return false
	}
	return true
}

func (ee *Events) quitListener(ev *tcell.EventKey) func(*Env) {
	ee.mutex.Lock()
	defer ee.mutex.Unlock()
	return ee.quit
}
