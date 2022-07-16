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

// Register provides a (wrapped) lines.Register fixture augmented with
// useful features for testing like firing an event or getting the
// current screen content as string.  Use *New* to create a new instance
// of Register.
// NOTE do not use a Register-instance concurrently.

// NOTE Register-fixture's Listen-method is non-blocking and starts the
// wrapped Register's event-loop in its own go-routine.  It is
// guaranteed that all methods of an Register-fixture-instance which
// trigger an event do not return before the event is processed and any
// view manipulations are printed to the screen.  NOTE the above is only
// true as long as you do not circumvent methods of this
// Register-fixture-type by calling them directly on the wrapped
// Register-instance.
type Register struct {
	*lines.Register
	lib        tcell.SimulationScreen
	mutex      *sync.Mutex
	reported   bool
	haveResize bool
	t          *gounit.T

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
		mutex: &sync.Mutex{}, Timeout: 100 * time.Millisecond}
	if len(max) > 0 {
		fx.Max = max[0]
	}
	return &fx
}

func (rg *Register) falsifyReported() {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	rg.reported = false
}

func (rg *Register) setReported() {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	rg.reported = true
}

func (rg *Register) hasBeenReported() bool {
	rg.mutex.Lock()
	defer rg.mutex.Unlock()
	return rg.reported
}

// SetNumberOfLines fires a resize event setting the screen lines to
// given number.  Note if an resize event listener is registered we can
// directly wait on returned channel.  SetNumberOfLines posts a resize
// event and returns after this event has been processed.
func (rg *Register) SetNumberOfLines(n int) {
	rg.t.GoT().Helper()
	if !rg.IsPolling() {
		rg.t.Fatal("fire key: not polling")
	}
	w, _ := rg.lib.Size()
	rg.lib.SetSize(w, n)
	rg.falsifyReported()
	rg.t.FatalOn(rg.lib.PostEvent(tcell.NewEventResize(w, n)))
	select {
	case <-rg.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("set number of lines: sync timed out")
	}
	if rg.hasBeenReported() {
		rg.decrementMaxEvents()
	}
}

// FireRuneEvent posts given run-key-press event and returns after this
// event has been processed.  Note modifier keys are ignored for
// rune-triggered key-events.
func (rg *Register) FireRuneEvent(r rune) {
	rg.t.GoT().Helper()
	if !rg.IsPolling() {
		rg.t.Fatal("fire rune: not polling")
	}
	rg.falsifyReported()
	rg.lib.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	select {
	case <-rg.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("fire rune: sync timed out")
	}
	if rg.hasBeenReported() {
		rg.decrementMaxEvents()
	}
}

// FireKeyEvent posts given special-key-press event and returns after
// this event has been processed.
func (rg *Register) FireKeyEvent(k tcell.Key, m ...tcell.ModMask) {
	rg.t.GoT().Helper()
	if !rg.IsPolling() {
		rg.t.Fatal("fire key: not polling")
	}
	rg.falsifyReported()
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
	if rg.hasBeenReported() {
		rg.decrementMaxEvents()
	}
}

// Listen posts the initial resize event and calls the wrapped
// register's Listen-method in a new go-routine.  Listen returns after
// the initial resize has completed.
func (rg *Register) Listen() {
	rg.t.GoT().Helper()
	rg.falsifyReported()
	err := rg.lib.PostEvent(tcell.NewEventResize(rg.lib.Size()))
	rg.t.FatalOn(err)
	go rg.Register.Listen()
	select {
	case <-rg.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("listen: sync timed out")
	}
	if rg.hasBeenReported() {
		rg.decrementMaxEvents()
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

// QuitListening stops wrapped Register's event loop.  This method does
// not return before Register.IsPolling() returns false.
func (rg *Register) QuitListening() {
	rg.LastScreen = rg.String()
	polling := rg.IsPolling()
	rg.Register.QuitListening()
	if !polling {
		return
	}
	select {
	case <-rg.Register.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("quit listening: sync timed out")
	}
}

// Resize wraps given listener for MaxEvent-maintenance before it is
// passed on to the wrapped view-*Register* property.
func (rg *Register) Resize(listener func(*lines.View)) {
	if listener == nil {
		if rg.haveResize {
			rg.haveResize = false
		}
		rg.Register.Resize(listener)
		return
	}
	if !rg.haveResize {
		rg.haveResize = true
	}
	rg.Register.Resize(rg.resizeWrapper(listener))
}

func (rg *Register) resizeWrapper(
	l func(*lines.View),
) func(*lines.View) {
	return func(v *lines.View) {
		l(v)
		rg.setReported()
	}
}

// Update wraps given listener for MaxEvent-maintenance before it is
// passed on to the wrapped *Register*'s Update method.  The later posts an
// event in case the listener is not nil; Update doesn't return before
// this event is processed.
func (rg *Register) Update(listener func(*lines.View)) error {
	rg.t.GoT().Helper()
	if !rg.IsPolling() {
		rg.t.Fatal("update: not polling")
	}
	if listener == nil {
		return rg.Register.Update(listener)
	}
	dropped := true
	l := func(v *lines.View) {
		dropped = false
		listener(v)
	}
	err := rg.Register.Update(l)
	if err != nil {
		return err
	}
	select {
	case <-rg.Register.Synced:
	case <-rg.t.Timeout(rg.Timeout):
		rg.t.Fatalf("update wrapper: sync timed out")
	}
	if !dropped {
		rg.decrementMaxEvents()
	}
	return nil
}

// Quit wraps given listener for MaxEvent-maintenance before it is
// passed on to the wrapped *Register* property.
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
func (rg *Register) Rune(r rune, listener func(*lines.View)) error {
	if listener == nil {
		return rg.Register.Rune(r, listener)
	}
	return rg.Register.Rune(r, rg.runeWrapper(listener))
}

func (rg *Register) runeWrapper(
	l func(*lines.View),
) func(*lines.View) {
	return func(v *lines.View) {
		l(v)
		rg.setReported()
	}
}

// Keyboard wraps given listener for MaxEvent-maintenance before its passed
// on to the wrapped Register instance.
func (rg *Register) Keyboard(
	listener func(*lines.View, rune, tcell.Key, tcell.ModMask),
) {
	if listener == nil {
		rg.Register.Keyboard(listener)
		return
	}
	rg.Register.Keyboard(rg.keyboardWrapper(listener))
}

func (rg *Register) keyboardWrapper(
	l func(*lines.View, rune, tcell.Key, tcell.ModMask),
) func(*lines.View, rune, tcell.Key, tcell.ModMask) {
	return func(v *lines.View, r rune, k tcell.Key, m tcell.ModMask) {
		l(v, r, k, m)
		rg.setReported()
	}
}

// Key wraps given listener for MaxEvent-maintenance before it is
// passed on to the wrapped view-*Register* property.
func (rg *Register) Key(
	k tcell.Key, m tcell.ModMask, listener lines.Listener,
) error {
	if listener == nil {
		return rg.Register.Key(k, m, listener)
	}
	return rg.Register.Key(k, m, rg.keyWrapper(listener))
}

func (rg *Register) keyWrapper(l lines.Listener) lines.Listener {
	return func(v *lines.View) {
		l(v)
		rg.setReported()
	}
}

func (rg *Register) decrementMaxEvents() {
	rg.Max--
	if rg.Max < 0 {
		rg.QuitListening()
		select {
		case <-rg.Register.Synced:
		case <-rg.t.Timeout(10 * time.Millisecond):
			rg.t.Fatalf("quit listening: sync timed out")
		}
		return
	}
}
