// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package mck mocks up a line.View for testing.
package fx

import (
	"github.com/gdamore/tcell/v2"
	"github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

// View embeds a *lines.View* and adds features for testing to remove
// noise from suite-test.
type View struct {
	*lines.View
	lib tcell.SimulationScreen
	t   *gounit.T

	// MaxEvents is the number of reported events after which the
	// event-loop a view-fixture is terminated.  MaxEvents is
	// decremented after each reported event.  I.e. events for which no
	// listener is registered are not counted.
	MaxEvents int

	// NextEventProcessed receives a message after each event-listener
	// call.  It is closed if the event-loop terminates after MaxEvents
	// many events have been reported to event-listeners.
	NextEventProcessed chan struct{}
}

// NewView creates a new lines.View test-fixture with additional
// features for testing.  If a positive number n is given the event-loop
// is automatically terminated after this amount of events have been
// reported.  Is no number of max-events given the event-loop stops
// after the first reported event.
func NewView(t *gounit.T, maxEvents ...int) *View {
	t.GoT().Helper()
	sim, lib, err := lines.NewSim()
	t.FatalOn(err)
	v := View{View: sim, lib: lib, t: t,
		NextEventProcessed: make(chan struct{}, 1)}
	sim.Register = &RegisterWrapper{ListenerRegister: sim.Register, vw: &v}
	if len(maxEvents) > 0 {
		v.MaxEvents = maxEvents[0]
	}
	return &v
}

func (v *View) FxRegister() *RegisterWrapper {
	return v.View.Register.(*RegisterWrapper)
}

// SetNumberOfLines sets the screen lines to given number.
func (v *View) SetNumberOfLines(n int) {
	v.t.GoT().Helper()
	w, _ := v.lib.Size()
	v.lib.SetSize(w, n)
	v.t.FatalOn(v.lib.PostEvent(tcell.NewEventResize(w, n)))
}

// FireRuneEvent dispatches given run-key-press event.  Note modifier
// keys are ignored for rune-triggered key-events.
func (v *View) FireRuneEvent(r rune) {
	v.lib.InjectKey(tcell.KeyRune, r, tcell.ModNone)
}

func (v *View) FireKeyEvent(k tcell.Key, m ...tcell.ModMask) {
	if len(m) == 0 {
		v.lib.InjectKey(k, 0, tcell.ModNone)
		return
	}
	v.lib.InjectKey(k, 0, m[0])
}

// Listen posts the initial resize event and starts the wrapped View's
// event-loop.
func (v *View) Listen() error {
	v.t.GoT().Helper()
	err := v.lib.PostEvent(tcell.NewEventResize(v.lib.Size()))
	v.t.FatalOn(err)
	if err := v.View.Listen(); err != nil {
		close(v.NextEventProcessed)
		return err
	}
	return nil
}

// Quit exits wrapped View's event loop and closes the
// *NextEventProcessed* channel.
func (v *View) Quit() {
	v.View.Quit()
	close(v.NextEventProcessed)
}

type RegisterWrapper struct {
	lines.ListenerRegister
	vw *View
}

// Rune wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (r *RegisterWrapper) Rune(listener func(*lines.View), rr ...rune) error {
	if listener == nil {
		return r.ListenerRegister.Rune(listener, rr...)
	}
	return r.ListenerRegister.Rune(r.runeWrapper(listener), rr...)
}

func (r *RegisterWrapper) runeWrapper(
	l func(*lines.View),
) func(*lines.View) {
	return func(v *lines.View) {
		l(v)
		r.informAboutProcessedEvent()
		r.decrementMaxEvents()
	}
}

func (r *RegisterWrapper) Runes(listener func(*lines.View, rune)) {
	if listener == nil {
		r.ListenerRegister.Runes(listener)
		return
	}
	r.ListenerRegister.Runes(r.runesWrapper(listener))
}

func (r *RegisterWrapper) runesWrapper(
	l func(*lines.View, rune),
) func(*lines.View, rune) {
	return func(v *lines.View, _r rune) {
		l(v, _r)
		r.informAboutProcessedEvent()
		r.decrementMaxEvents()
	}
}

// Key wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (r *RegisterWrapper) Key(
	listener func(*lines.View, tcell.ModMask), kk ...tcell.Key,
) error {
	if listener == nil {
		return r.ListenerRegister.Key(listener, kk...)
	}
	return r.ListenerRegister.Key(r.keyWrapper(listener), kk...)
}

func (r *RegisterWrapper) keyWrapper(
	l func(*lines.View, tcell.ModMask),
) func(*lines.View, tcell.ModMask) {
	return func(v *lines.View, m tcell.ModMask) {
		l(v, m)
		r.informAboutProcessedEvent()
		r.decrementMaxEvents()
	}
}

func (r *RegisterWrapper) resizeWrapper(
	l func(*lines.View),
) func(*lines.View) {
	return func(v *lines.View) {
		l(v)
		r.informAboutProcessedEvent()
		r.decrementMaxEvents()
	}
}

// Resize wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (r *RegisterWrapper) Resize(listener func(*lines.View)) {
	if listener == nil {
		r.ListenerRegister.Resize(listener)
		return
	}
	r.ListenerRegister.Resize(r.resizeWrapper(listener))
}

func (r *RegisterWrapper) quitWrapper(l func()) func() {
	return func() {
		l()
		r.vw.MaxEvents--
		close(r.vw.NextEventProcessed)
	}
}

// Quit wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (r *RegisterWrapper) Quit(listener func()) {
	if listener == nil {
		r.ListenerRegister.Quit(listener)
		return
	}
	r.ListenerRegister.Quit(r.quitWrapper(listener))
}

func (r *RegisterWrapper) decrementMaxEvents() {
	r.vw.MaxEvents--
	if r.vw.MaxEvents >= 0 {
		return
	}
	r.vw.Quit()
}

func (r *RegisterWrapper) informAboutProcessedEvent() {
	select {
	case r.vw.NextEventProcessed <- struct{}{}:
	default:
	}
}
