// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Register allows to register for user-input events and posting
// update-events.  Registering of events is concurrency save.
type Register struct {
	view      *View
	keys      *sync.Mutex
	kk        map[tcell.Key]func(*View, tcell.ModMask)
	runes     *sync.Mutex
	rr        map[rune]func(*View)
	other     *sync.Mutex
	resize    func(*View)
	quit      func()
	allRunes  func(*View, rune)
	isPolling bool

	// Synced sends a message after a the screen synchronization
	// following a reported event.
	Synced chan bool

	// Ev provides the currently reported tcell-event.
	Ev tcell.Event
}

// IsPolling returns true if listener register is polling in the event
// loop.
func (rg *Register) IsPolling() bool { return rg.isPolling }

// Listen blocks and starts polling from the event loop reporting
// received events to registered listeners.  Listen returns if either a
// quit-event was received ('q', ctrl-c, ctrl-d input) or QuitListening
// was called.
func (rg *Register) Listen() {
	rg.isPolling = true
	for {
		ev := rg.view.lib.PollEvent()

		select {
		case <-rg.Synced:
		default:
		}

		switch ev := ev.(type) {
		case nil: // event-loop ended
			return
		case *tcell.EventResize:
			if rg.view.resize() {
				rg.report(ev)
			}
			rg.view.ensureSynced(false)
			rg.Synced <- true
		default:
			quit := rg.report(ev)
			if quit {
				rg.QuitListening()
				return
			}
			rg.view.ensureSynced(true)
			rg.Synced <- true
		}
	}
}

// Resize registers given listener for the resize event.  Note starting
// the event-loop by calling *Listen* will trigger a mandatory initial
// resize event.
func (rg *Register) Resize(listener func(*View)) {
	rg.other.Lock()
	defer rg.other.Unlock()
	rg.resize = listener
}

// Quit registers given listener for the quit event which is triggered
// by 'r'-rune, ctrl-c and ctrl-d.
func (rg *Register) Quit(listener func()) {
	rg.other.Lock()
	defer rg.other.Unlock()
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

// RegisterErr is returned by Register.Rune and Register.Key if a
// listener is registered for an already registered rune/key-event.
var RegisterErr = errors.New("event listener overwrites existing")

// Rune registers provided listener for each of the given rune's input
// events.  Rune fails if for one of the runes already a listener is
// registered.  In the later case the listener isn't registered for any
// of the given runes.  If the listener is nil given runes are
// unregistered.
func (rg *Register) Rune(listener func(*View), rr ...rune) error {
	rg.runes.Lock()
	defer rg.runes.Unlock()

	if listener == nil {
		for _, _r := range rr {
			delete(rg.rr, _r)
		}
		return nil
	}

	for _, _r := range rr {
		if _, ok := rg.rr[_r]; ok {
			return RegisterErr
		}
		if _r == 'q' {
			return RegisterErr
		}
	}

	for _, _r := range rr {
		rg.rr[_r] = listener
	}
	return nil
}

// Runes suppresses all registers Rune-events and reports rune input
// events to given listener only until a non-rune is inserted.
// TODO: add to the listener a cancel argument and to the Runes-method a
// tcell.Key argument defining the key which ends the Runes
// registration, i.e. calls listener for a last time with cancel set to
// true.
func (rg *Register) Runes(listener func(*View, rune)) {
	rg.other.Lock()
	defer rg.other.Unlock()
	rg.allRunes = listener
}

// Key registers provided listener for each of the given key's input
// events.  Key fails if for one of the keys already a listener is
// registered.  In the later case the listener isn't registered for any
// of the given keys.  If the listener is nil given keys are
// unregistered.
func (rg *Register) Key(
	listener func(*View, tcell.ModMask), kk ...tcell.Key,
) error {
	rg.keys.Lock()
	defer rg.keys.Unlock()

	if listener == nil {
		for _, k := range kk {
			delete(rg.kk, k)
		}
		return nil
	}

	for _, k := range kk {
		if _, ok := rg.kk[k]; ok {
			return RegisterErr
		}
		if k == tcell.KeyCtrlC || k == tcell.KeyCtrlD {
			return RegisterErr
		}
	}

	for _, k := range kk {
		rg.kk[k] = listener
	}
	return nil
}

// QuitListening ends the event-loop, i.e. IsPolling will be false.
func (rg *Register) QuitListening() {
	rg.isPolling = false
	rg.view.lib.Fini()
	close(rg.Synced)
}

func (rg *Register) report(ev tcell.Event) (quit bool) {
	rg.Ev = ev
	if rg.view.ToSmall() {
		if ev, ok := ev.(*tcell.EventKey); ok {
			return rg.reportQuit(ev)
		}
		return false
	}
	switch ev := ev.(type) {
	case *tcell.EventResize:
		rg.reportResize(rg.view) // TODO: remove view arg
	case *tcell.EventKey:
		if rg.reportQuit(ev) {
			return true
		}
		if ev.Key() == tcell.KeyRune {
			rg.reportRune(rg.view, ev.Rune()) // TODO: remove view arg
		} else {
			rg.reportKey(rg.view, ev.Key(), ev.Modifiers()) // TODO: remove view arg
		}
	case *updateEvent:
		ev.listener(rg.view)
	}
	return false
}

func (rg *Register) reportRune(v *View, r rune) {
	rg.other.Lock()
	if rg.allRunes != nil {
		rg.allRunes(v, r)
		rg.other.Unlock()
		return
	}
	rg.other.Unlock()
	rg.runes.Lock()
	defer rg.runes.Unlock()
	if _, ok := rg.rr[r]; !ok {
		return
	}
	rg.rr[r](v)
}

func (rg *Register) reportKey(v *View, k tcell.Key, m tcell.ModMask) {
	rg.keys.Lock()
	defer rg.keys.Unlock()
	if _, ok := rg.kk[k]; !ok {
		return
	}
	rg.kk[k](v, m)
}

// TODO: refac
func (rg *Register) reportResize(v *View) {
	rg.other.Lock()
	defer rg.other.Unlock()
	if rg.resize == nil {
		return
	}
	rg.resize(v)
}

func (rg *Register) reportQuit(ev *tcell.EventKey) bool {
	rg.other.Lock()
	defer rg.other.Unlock()
	if ev.Key() == tcell.KeyRune && ev.Rune() != 'q' {
		return false
	}
	if ev.Key() != tcell.KeyRune && ev.Key() != tcell.KeyCtrlC &&
		ev.Key() != tcell.KeyCtrlD {
		return false
	}
	if rg.quit == nil {
		return true
	}
	rg.quit()
	return true
}
