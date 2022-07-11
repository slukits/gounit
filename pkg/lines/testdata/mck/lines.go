// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package mck mocks up a line.View for testing.
package mck

import (
	"github.com/gdamore/tcell/v2"
	"github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

// View embeds a *lines.View* and adds features for testing to remove
// noise from suite-test.
type View struct {
	*lines.View
	sim tcell.SimulationScreen
	t   *gounit.T

	// MaxEvents is the number of events after which the event-loop is
	// terminated.  MaxEvents is decremented after each event-listener
	// call.  I.e. events for which no listener is registered are not
	// counted.
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
	v := View{View: sim, sim: lib, t: t,
		NextEventProcessed: make(chan struct{}, 1)}
	if len(maxEvents) > 0 {
		v.MaxEvents = maxEvents[0]
	}
	return &v
}

// SetNumberOfLines sets the screen lines to given number.
func (v *View) SetNumberOfLines(n int) {
	v.t.GoT().Helper()
	w, _ := v.sim.Size()
	v.sim.SetSize(w, n)
	v.t.FatalOn(v.sim.PostEvent(tcell.NewEventResize(w, n)))
}

// FireRuneEvent dispatches given run-key-press event.  Note modifier
// keys are ignored for rune-triggered key-events.
func (v *View) FireRuneEvent(r rune) {
	v.sim.InjectKey(tcell.KeyRune, r, tcell.ModNone)
}

// Listen posts the initial resize event and starts the wrapped View's
// event-loop.
func (v *View) Listen() error {
	v.t.GoT().Helper()
	v.wrapListeners()
	err := v.sim.PostEvent(tcell.NewEventResize(v.sim.Size()))
	v.t.FatalOn(err)
	if err := v.View.Listen(); err != nil {
		close(v.NextEventProcessed)
		return err
	}
	return nil
}

func (v *View) wrapListeners() {
	if v.View.Listeners.Resize != nil {
		v.View.Listeners.Resize = v.resizeWrapper(
			v.View.Listeners.Resize)
	}
	if v.View.Listeners.Quit != nil {
		v.View.Listeners.Quit = v.quitWrapper(
			v.View.Listeners.Quit)
	}
}

func (v *View) resizeWrapper(l func(*lines.View)) func(*lines.View) {
	return func(lv *lines.View) {
		l(lv)
		v.informAboutProcessedEvent()
		v.decrementMaxEvents()
	}
}

func (v *View) quitWrapper(l func()) func() {
	return func() {
		l()
		v.MaxEvents--
		close(v.NextEventProcessed)
	}
}

func (v *View) decrementMaxEvents() {
	v.MaxEvents--
	if v.MaxEvents >= 0 {
		return
	}
	v.Quit()
}

func (v *View) informAboutProcessedEvent() {
	select {
	case v.NextEventProcessed <- struct{}{}:
	default:
	}
}

// Quit exits wrapped View's event loop and closes the
// *NextEventProcessed* channel.
func (v *View) Quit() {
	v.View.Quit()
	close(v.NextEventProcessed)
}
