// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gdamore/tcell/v2"
)

// Listener is the most common type of listener: a callback provided
// with the view-instance.
type Listener = func(*Env)

// KBListener is a keyboard callback instance provided with the view and
// all information about the received key event as reported by tcell.
type KBListener = func(*Env, rune, tcell.Key, tcell.ModMask)

// Listeners resembles a concurrency save set of registered event
// listeners mapped to their events.
type Listeners struct {
	mutex      *sync.Mutex
	isQuitting func(k tcell.Key, r rune) bool
	kk         map[tcell.ModMask]map[tcell.Key]Listener
	rr         map[rune]Listener
	kb         KBListener
}

// IsQuitter provides the information which runes/keys may not be used
// for listener registration since they are used for quitting.
type IsQuitter interface {
	// RuneQuits returns true iff given rune (event) quits event loop
	// listening.
	RuneQuits(rune) bool
	// KeyQuits returns true iff given key (event without modifier)
	// quits event loop listening.
	KeyQuits(tcell.Key) bool
}

// NewListener creates a new listener instance whereas given is-quitter
// defines the runes and keys which are registered for quitting the
// listening to the event-loop.  If IsQuitter is nil DefaultFeatures is
// used.  Note Features implements IsQuitter interface which is need to
// produces corresponding errors when registering rune/key listeners
// with the same runes/keys that are registered for quitting.
func NewListeners(ff IsQuitter) *Listeners {
	return &Listeners{
		isQuitting: isQuitterClosure(ff),
		mutex:      &sync.Mutex{},
		rr:         map[rune]func(*Env){},
		kk:         map[tcell.ModMask]map[tcell.Key]func(*Env){},
	}
}

// ErrEvents is the general type for errors at listener registration.
var ErrEvents = errors.New("add event")

// ErrZeroRune for attempting to register for the zero-rune.
var ErrZeroRune = fmt.Errorf("%w: can't register zero-rune", ErrEvents)

// ErrZeroKey for attempting to register for the zero-key
var ErrZeroKey = fmt.Errorf("%w: can't register zero-key", ErrEvents)

// ErrQuit for attempting to register for a key/rune which is associated
// with the quit-feature.
var ErrQuit = fmt.Errorf("%w: associated with quit event", ErrEvents)

// ErrExists for attempting to register for a key/rune which is already
// registered.
var ErrExists = fmt.Errorf("%w: already registered", ErrEvents)

// Rune registers provided listener for given rune respectively removes
// the registration for given rune if the listener is nil.  Rune fails
// if already a listener is registered for given rune or if the zero
// rune is given or if given rune is associated with the quit-feature.
// NOTE use *Quit* at an Register-instance to receive the Quit-event.
func (kk *Listeners) Rune(r rune, l Listener) error {
	kk.mutex.Lock()
	defer kk.mutex.Unlock()

	if l == nil {
		delete(kk.rr, r)
		return nil
	}

	if r == rune(0) {
		return ErrZeroRune
	}
	if kk.isQuitting(0, r) {
		return ErrQuit
	}
	if _, ok := kk.rr[r]; ok {
		return fmt.Errorf("%w: %c", ErrExists, r)
	}

	kk.rr[r] = l
	return nil
}

// Keyboard listener shadows all other rune/key listeners except for the
// quite event until it is removed by Keys.Keyboard(nil).
func (kk *Listeners) Keyboard(l KBListener) {
	kk.mutex.Lock()
	defer kk.mutex.Unlock()

	kk.kb = l
}

func (kk *Listeners) KBListener() KBListener {
	kk.mutex.Lock()
	defer kk.mutex.Unlock()

	return kk.kb
}

// Key registers provided listener for given key/mode combination
// respectively removes the registration for given key/mode if the
// listener is nil.  Key fails if already a listener is registered for
// given key/mode or if the zero key is given or if given key is
// associated with the quit-feature.  NOTE use *Quit* at an
// Register-instance to receive the Quit-event.
func (kk *Listeners) Key(k tcell.Key, m tcell.ModMask, l Listener) error {
	kk.mutex.Lock()
	defer kk.mutex.Unlock()

	if l == nil {
		if kk.kk[m] != nil {
			delete(kk.kk[m], k)
		}
		return nil
	}

	if k == tcell.KeyNUL {
		return ErrZeroKey
	}
	if kk.isQuitting(k, 0) {
		return ErrQuit
	}

	if kk.kk[m] == nil {
		kk.kk[m] = map[tcell.Key]Listener{k: l}
		return nil
	}

	if _, ok := kk.kk[m][k]; ok {
		return ErrExists
	}

	kk.kk[m][k] = l
	return nil
}

// KeyListenerOf returns the listener registered for given key/mode
// combination.  The second return value is false if no listener is
// registered for given key.
func (kk *Listeners) KeyListenerOf(
	k tcell.Key, m tcell.ModMask,
) (Listener, bool) {
	kk.mutex.Lock()
	defer kk.mutex.Unlock()

	if _, ok := kk.kk[m]; !ok {
		return nil, false
	}
	l, ok := kk.kk[m][k]
	return l, ok
}

// RuneListenerOf returns the listener registered for given rune.  The
// second return value is false if no listener is registered for given
// rune.
func (kk *Listeners) RuneListenerOf(r rune) (Listener, bool) {
	kk.mutex.Lock()
	defer kk.mutex.Unlock()

	l, ok := kk.rr[r]
	return l, ok
}

// HasKBListener is true if a KBListener is registered.
func (kk *Listeners) HasKBListener() bool {
	kk.mutex.Lock()
	defer kk.mutex.Unlock()

	return kk.kb != nil
}
