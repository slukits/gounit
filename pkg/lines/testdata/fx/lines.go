// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package mck mocks up a line.View for testing.
package fx

import (
	"strings"

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
	// many events have been reported to event-listeners.  NOTE it is
	// guaranteed that the message is sent *after* all lines-updates have
	// made it to the screen.
	NextEventProcessed chan struct{}

	// LastScreen provides the screen content right before quitting.
	// NOTE it is guaranteed that that this snapshot is taken *after*
	// all lines-updates have made it to the screen
	LastScreen string
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

// SetNumberOfLines sets the screen lines to given number.  Note if an
// resize event listener is registered we can directly wait on returned
// channel.
func (v *View) SetNumberOfLines(n int) chan struct{} {
	v.t.GoT().Helper()
	w, _ := v.lib.Size()
	v.lib.SetSize(w, n)
	v.t.FatalOn(v.lib.PostEvent(tcell.NewEventResize(w, n)))
	return v.NextEventProcessed
}

// FireRuneEvent dispatches given run-key-press event.  Note modifier
// keys are ignored for rune-triggered key-events.  Note if an event
// listener is registered for this rune we can directly wait on returned
// channel.
func (v *View) FireRuneEvent(r rune) chan struct{} {
	v.lib.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	return v.NextEventProcessed
}

// FireKeyEvent dispatches given special-key-press event.  Note if an
// event listener is registered for this key we can directly wait on
// returned channel.
func (v *View) FireKeyEvent(k tcell.Key, m ...tcell.ModMask) chan struct{} {
	if len(m) == 0 {
		v.lib.InjectKey(k, 0, tcell.ModNone)
	} else {
		v.lib.InjectKey(k, 0, m[0])
	}
	return v.NextEventProcessed
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

// String returns the test-screen's content as string with line breaks
// where a new screen line starts.  Empty lines at the end of the screen
// are not returned and empty cells at the end of a line are trimmed.
func (v *View) String() string {
	b, w, h := v.lib.GetContents()
	sb := &strings.Builder{}
	for i := 0; i < h; i++ {
		line := ""
		for j := 0; j < w; j++ {
			cell := b[cellIdx(j, i, w)]
			if len(cell.Runes) == 0 {
				continue
			}
			line += string(cell.Runes[0])
		}
		if len(strings.TrimSpace(line)) == 0 {
			sb.WriteString("\n")
			continue
		}
		sb.WriteString(strings.TrimRight(
			line, " \t\r") + "\n")
	}
	return strings.TrimLeft(
		strings.TrimRight(sb.String(), " \t\r\n"), "\n")
}

func cellIdx(x, y, w int) int {
	if x == 0 {
		return y * w
	}
	if y == 0 {
		return x
	}
	return y*w + x
}

// Quit exits wrapped View's event loop and closes the
// *NextEventProcessed* channel.
func (v *View) Quit() {
	v.LastScreen = v.String()
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
		go func() {
			<-v.Synced
			r.decrementMaxEvents()
		}()
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
		go func() {
			<-v.Synced
			r.decrementMaxEvents()
		}()
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
		go func() {
			<-v.Synced
			r.decrementMaxEvents()
		}()
	}
}

func (r *RegisterWrapper) resizeWrapper(
	l func(*lines.View),
) func(*lines.View) {
	return func(v *lines.View) {
		l(v)
		go func() {
			<-v.Synced
			r.decrementMaxEvents()
		}()
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
	if r.vw.MaxEvents < 0 {
		r.vw.Quit()
		<-r.vw.View.Synced // wait for closing to finish
		return
	}
	r.informAboutProcessedEvent()
}

func (r *RegisterWrapper) informAboutProcessedEvent() {
	select {
	case r.vw.NextEventProcessed <- struct{}{}:
	default:
	}
}
