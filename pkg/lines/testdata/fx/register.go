// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fx

import (
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

// Register provides a (wrapped) lines.Register fixture augmented with
// useful features for testing like firing an event or getting the
// current screen content as string.  Use *New* to create a new instance
// of Register.
type Register struct {
	*lines.Register
	lib tcell.SimulationScreen
	t   *gounit.T

	// Max is the number of reported events after which the
	// event-loop of a register-fixture is terminated.  Max is
	// decremented after each reported event.  I.e. events for which no
	// listener is registered are not counted.
	Max int

	// NextEventProcessed receives a message after each event-listener
	// call.  It is closed if the event-loop terminates after MaxEvents
	// many events have been reported to event-listeners.  NOTE it is
	// guaranteed that the message is sent *after* all lines-updates have
	// made it to the screen.
	NextEventProcessed chan struct{}

	// LastScreen provides the screen content right before quitting
	// listening.  NOTE it is guaranteed that that this snapshot is
	// taken *after* all lines-updates have made it to the screen.
	LastScreen string
}

// New creates a new Register test-fixture with additional features for
// testing.  If a positive number n is given the event-loop is
// automatically terminated after this amount of events have been
// reported.  Is no number of max-events given the event-loop stops
// after the first reported event.
func New(t *gounit.T, max ...int) *Register {
	t.GoT().Helper()
	reg, lib, err := lines.Sim()
	t.FatalOn(err)
	fx := Register{Register: reg, lib: lib, t: t,
		NextEventProcessed: make(chan struct{}, 1)}
	if len(max) > 0 {
		fx.Max = max[0]
	}
	return &fx
}

// SetNumberOfLines fires a resize event setting the screen lines to
// given number.  Note if an resize event listener is registered we can
// directly wait on returned channel.
func (rg *Register) SetNumberOfLines(n int) chan struct{} {
	rg.t.GoT().Helper()
	w, _ := rg.lib.Size()
	rg.lib.SetSize(w, n)
	rg.t.FatalOn(rg.lib.PostEvent(tcell.NewEventResize(w, n)))
	return rg.NextEventProcessed
}

// FireRuneEvent dispatches given run-key-press event.  Note modifier
// keys are ignored for rune-triggered key-events.  Note if an event
// listener is registered for this rune we can directly wait on returned
// channel.
func (rg *Register) FireRuneEvent(r rune) chan struct{} {
	rg.lib.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	return rg.NextEventProcessed
}

// FireKeyEvent dispatches given special-key-press event.  Note if an
// event listener is registered for this key we can directly wait on
// returned channel.
func (rg *Register) FireKeyEvent(
	k tcell.Key, m ...tcell.ModMask,
) chan struct{} {
	if len(m) == 0 {
		rg.lib.InjectKey(k, 0, tcell.ModNone)
	} else {
		rg.lib.InjectKey(k, 0, m[0])
	}
	return rg.NextEventProcessed
}

// Listen posts the initial resize event and calls the wrapped
// register's Listen-method.
func (rg *Register) Listen() {
	rg.t.GoT().Helper()
	err := rg.lib.PostEvent(tcell.NewEventResize(rg.lib.Size()))
	rg.t.FatalOn(err)
	rg.Register.Listen()
	close(rg.NextEventProcessed)
}

// String returns the test-screen's content as string with line breaks
// where a new screen line starts.  Empty lines at the end of the screen
// are not returned and empty cells at the end of a line are trimmed.
// I.e.
// +-------------+
// |             |
// |   content   |   => "content"
// |             |
// +-------------+
func (rg *Register) String() string {
	b, w, h := rg.lib.GetContents()
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

// QuitListening stops wrapped Register's event loop and closes the
// *NextEventProcessed* channel.
func (rg *Register) QuitListening() {
	rg.LastScreen = rg.String()
	rg.Register.QuitListening()
}

// Resize wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (rg *Register) Resize(listener func(*lines.View)) {
	if listener == nil {
		rg.Register.Resize(listener)
		return
	}
	rg.Register.Resize(rg.resizeWrapper(listener))
}

func (rg *Register) resizeWrapper(
	l func(*lines.View),
) func(*lines.View) {
	return func(v *lines.View) {
		l(v)
		go func() {
			<-rg.Register.Synced
			rg.decrementMaxEvents()
		}()
	}
}

// Update wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (rg *Register) Update(listener func(*lines.View)) error {
	if listener == nil {
		return rg.Register.Update(listener)
	}
	return rg.Register.Update(rg.updateWrapper(listener))
}

func (rg *Register) updateWrapper(
	l func(*lines.View),
) func(*lines.View) {
	return func(v *lines.View) {
		l(v)
		go func() {
			<-rg.Register.Synced
			rg.decrementMaxEvents()
		}()
	}
}

// Quit wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (rg *Register) Quit(listener func()) {
	if listener == nil {
		rg.Register.Quit(listener)
		return
	}
	rg.Register.Quit(rg.quitWrapper(listener))
}

func (rg *Register) quitWrapper(l func()) func() {
	return func() {
		l()
		rg.Max--
	}
}

// Rune wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (rg *Register) Rune(listener func(*lines.View), rr ...rune) error {
	if listener == nil {
		return rg.Register.Rune(listener, rr...)
	}
	return rg.Register.Rune(rg.runeWrapper(listener), rr...)
}

func (rg *Register) runeWrapper(
	l func(*lines.View),
) func(*lines.View) {
	return func(v *lines.View) {
		l(v)
		go func() {
			<-rg.Register.Synced
			rg.decrementMaxEvents()
		}()
	}
}

// Runes wraps given listener for MaxEvent-maintenance before its passed
// on to wrapped Register instance.
func (rg *Register) Runes(listener func(*lines.View, rune)) {
	if listener == nil {
		rg.Register.Runes(listener)
		return
	}
	rg.Register.Runes(rg.runesWrapper(listener))
}

func (rg *Register) runesWrapper(
	l func(*lines.View, rune),
) func(*lines.View, rune) {
	return func(v *lines.View, r rune) {
		l(v, r)
		go func() {
			<-rg.Register.Synced
			rg.decrementMaxEvents()
		}()
	}
}

// Key wraps given listener for MaxEvent-maintenance before it is
// passed on to wrapped view-*Register* property.
func (rg *Register) Key(
	listener func(*lines.View, tcell.ModMask), kk ...tcell.Key,
) error {
	if listener == nil {
		return rg.Register.Key(listener, kk...)
	}
	return rg.Register.Key(rg.keyWrapper(listener), kk...)
}

func (rg *Register) keyWrapper(
	l func(*lines.View, tcell.ModMask),
) func(*lines.View, tcell.ModMask) {
	return func(v *lines.View, m tcell.ModMask) {
		l(v, m)
		go func() {
			<-rg.Register.Synced
			rg.decrementMaxEvents()
		}()
	}
}

func (rg *Register) decrementMaxEvents() {
	rg.Max--
	if rg.Max < 0 {
		rg.QuitListening()
		<-rg.Register.Synced // wait for closing to finish
		return
	}
	rg.informAboutProcessedEvent()
}

func (rg *Register) informAboutProcessedEvent() {
	select {
	case rg.NextEventProcessed <- struct{}{}:
	default:
	}
}
