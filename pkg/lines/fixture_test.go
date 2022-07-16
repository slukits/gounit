// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
)

// Events provides a (wrapped) lines.Events fixture augmented with
// useful features for testing like firing an event or getting the
// current screen content as string.  Use *New* to create a new instance
// of Events.
// NOTE do not use an Events-instance concurrently.
// NOTE Events-fixture's Listen-method is non-blocking and starts the
// wrapped Events's event-loop in its own go-routine.  It is
// guaranteed that all methods of an Events-fixture-instance which
// trigger an event do not return before the event is processed and any
// view manipulations are printed to the screen.
// NOTE the above is only true as long as you do not circumvent methods
// of this Events-fixture-type by calling them directly on the wrapped
// Events-instance.
type Events struct {
	*lines.Events
	lib           tcell.SimulationScreen
	mutex         *sync.Mutex
	reported      bool
	haveResize    bool
	autoTerminate bool
	t             *gounit.T

	// Max is the number of reported events after which the
	// event-loop of a register-fixture is terminated.  Max is
	// decremented after each reported event.  I.e. events for which no
	// listener is registered are not counted.
	Max int

	// LastScreen provides the screen content right before quitting
	// listening.  NOTE it is guaranteed that that this snapshot is
	// taken *after* all lines-updates have made it to the screen.
	LastScreen string

	// Timeout defines how long an event-triggering method waits for the
	// event to be processed.  It defaults to 100ms.
	Timeout time.Duration
}

func decrement(ee *Events) func() {
	return func() {
		ee.Max--
	}
}

// New creates a new Register test-fixture with additional features for
// testing.  If a positive number n is given the event-loop is
// automatically terminated after this amount of events have been
// reported.  Is a negative number given the loop will not be stopped
// automatically.  Is no number of max-events given the event-loop stops
// after the first reported event.
func New(t *gounit.T, max ...int) *Events {
	t.GoT().Helper()
	reg, lib, err := lines.Sim()
	t.FatalOn(err)
	fx := Events{Events: reg, lib: lib, t: t,
		mutex: &sync.Mutex{}, Timeout: 100 * time.Millisecond}
	if len(max) == 0 {
		fx.autoTerminate = true
		fx.Reported(decrement(&fx))
	}
	if len(max) > 0 {
		if max[0] >= 0 {
			fx.autoTerminate = true
			fx.Reported(decrement(&fx))
		}
		fx.Max = max[0]
	}
	return &fx
}

// SetNumberOfLines fires a resize event setting the screen lines to
// given number.  Note if an resize event listener is registered we can
// directly wait on returned channel.  SetNumberOfLines posts a resize
// event and returns after this event has been processed.
func (rg *Events) SetNumberOfLines(n int) {
	rg.t.GoT().Helper()
	if !rg.IsPolling() {
		rg.t.Fatal("fire key: not polling")
	}
	w, _ := rg.lib.Size()
	rg.lib.SetSize(w, n)
	rg.t.FatalOn(rg.lib.PostEvent(tcell.NewEventResize(w, n)))
	select {
	case <-rg.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("set number of lines: sync timed out")
	}
	rg.checkTermination()
}

// FireRuneEvent posts given run-key-press event and returns after this
// event has been processed.  Note modifier keys are ignored for
// rune-triggered key-events.
func (rg *Events) FireRuneEvent(r rune) {
	rg.t.GoT().Helper()
	if !rg.IsPolling() {
		rg.t.Fatal("fire rune: not polling")
	}
	rg.lib.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	select {
	case <-rg.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("fire rune: sync timed out")
	}
	rg.checkTermination()
}

// FireKeyEvent posts given special-key-press event and returns after
// this event has been processed.
func (rg *Events) FireKeyEvent(k tcell.Key, m ...tcell.ModMask) {
	rg.t.GoT().Helper()
	if !rg.IsPolling() {
		rg.t.Fatal("fire key: not polling")
	}
	if len(m) == 0 {
		rg.lib.InjectKey(k, 0, tcell.ModNone)
	} else {
		rg.lib.InjectKey(k, 0, m[0])
	}
	select {
	case <-rg.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("fire key: sync timed out")
	}
	rg.checkTermination()
}

// Listen posts the initial resize event and calls the wrapped
// register's Listen-method in a new go-routine.  Listen returns after
// the initial resize has completed.
func (rg *Events) Listen() {
	rg.t.GoT().Helper()
	err := rg.lib.PostEvent(tcell.NewEventResize(rg.lib.Size()))
	rg.t.FatalOn(err)
	go rg.Events.Listen()
	select {
	case <-rg.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("listen: sync timed out")
	}
	rg.checkTermination()
}

func (ee *Events) checkTermination() {
	if !ee.autoTerminate {
		return
	}
	if ee.Max < 0 {
		ee.QuitListening()
		select {
		case <-ee.Synced:
		case <-ee.t.Timeout(ee.Timeout):
			ee.t.Fatalf("quit listening: sync timed out")
		}
	}
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
func (rg *Events) String() string {
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

// QuitListening stops wrapped Register's event loop.  This method does
// not return before  Events.IsPolling() returns false.
func (rg *Events) QuitListening() {
	if !rg.IsPolling() {
		return
	}
	rg.LastScreen = rg.String()
	rg.Events.QuitListening()
	select {
	case <-rg.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("quit listening: sync timed out")
	}
}

// Update wraps given listener for MaxEvent-maintenance before it is
// passed on to the wrapped *Register*'s Update method.  The later posts an
// event in case the listener is not nil; Update doesn't return before
// this event is processed.
func (rg *Events) Update(listener func(*lines.View)) error {
	rg.t.GoT().Helper()
	if !rg.IsPolling() {
		rg.t.Fatal("update: not polling")
	}
	if listener == nil {
		return rg.Events.Update(listener)
	}
	err := rg.Events.Update(listener)
	if err == nil {
		select {
		case <-rg.Synced:
		case <-rg.t.Timeout(rg.Timeout):
			rg.t.Fatalf("update wrapper: sync timed out")
		}
		rg.checkTermination()
	}
	return err
}
