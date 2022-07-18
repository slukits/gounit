// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gdamore/tcell/v2"
)

// Testing augments lines.Events instance created by *Test* with useful
// features for testing like firing an event or getting the current
// screen content as string.
// NOTE do not use an Events/Testing-instances concurrently.
// NOTE Events.Listen-method becomes non-blocking and starts event-loop
// polling in its own go-routine.
// NOTE all event triggering methods start event-listening if it is not
// already started.
// NOTE It is guaranteed that all methods of an Events/Testing-instances
// which trigger an event do not return before the event is processed
// and any view manipulations are printed to the screen.  This holds
// also ture if you call an event triggering method within a listener
// callback, e.g.
//
// func TestTest(t *testing.T) {
//     ee := lines.Test(t).T().SetMax(3)
//     // Update is event triggering, i.e. starts event loop
//     ee.Update(func(e *lines.Env) {
//         e.LL().Line(0).Set("42")
//         // QuitListening is also event-triggering
//         e.EE.QuitListening()
//     })
//     // here it is guaranteed that all three events: initial resize,
//     // update and quitting are processed with all their screen output
//     if ee.IsListening() {
//         t.Error("expect listening to have stopped")
//     }
//     if ee.T().LastScreen != "42" {
//         t.Errorf("expected 42 on screen; got %s", ee.T().LastScreen)
//     }
// }
//
type Testing struct {
	ee            *Events
	lib           tcell.SimulationScreen
	autoTerminate bool
	mutex         *sync.Mutex
	waitStack     []string
	waiting       bool
	t             *testing.T

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

func decrement(ee *Testing) func() {
	return func() {
		ee.Max--
	}
}

// Test creates a new Events-test-fixture with additional features for
// testing.  *max* defaults to 1, i.e. after *Listen* was called the
// event-loop stops automatically after the first reported event.  Is
// *max* negative listening doesn't stop automatically.  *Max*
// decrements after each reported event.
// Timeout defaults to 100ms; i.e. how long the fixture waits during an
// event generating operation for the event being fully processed.  E.g.
//
//     func TestTest(t *testing.T) {
//         // stop listening after two reported events
//         ee, tt, resize := lines.Test(t, 2), 0
//         ee.Resize(func(e *lines.Env) {
//             e.LL().Line(resize).Setf("%d", e.Len())
//             resize++
//         })
//         // if an event-creating operation is used and the listening
//         // has not been started it is started automatically and the
//         // initial resize is fired.
//         tt.SetNumberOfLines(42)
//         // thus above line is the second reported event, i.e. the
//         // event listening is automatically terminated.
//         if !ee.IsListening() {
//             t.Error("expected stop listening after " +
//                 "2 reported events")
//         }
//         if !tt.Max != 0 {
//             t.Error("expected *Max* to be counted down " +
//                 "to 0; got: %d", ee.T().Max)
//         }
//         if tt.LastScreen != "25\n42" {
//             t.Errorf("expected screen 25\n42" +
//                 "got %s", tt.LastScreen)
//         }
//     }
//
// Event generating operations on the test-fixture are *Listen*,
// *SetNumberOfLines*, *FireKeyEvent*, *FireRuneEvent*; on the
// Events-instance: *QuitListening* and *Update*.
func Test(t *testing.T, max ...int) (*Events, *Testing) {
	t.Helper()
	ee, lib, err := Sim()
	if err != nil {
		t.Fatalf("test: init sim: %v", err)
	}
	ee.t = &Testing{ee: ee, lib: lib, t: t,
		Timeout: 200 * time.Millisecond,
		mutex:   &sync.Mutex{}}
	switch len(max) {
	case 0:
		ee.t.SetMax(1)
	default:
		ee.t.SetMax(max[0])
	}
	return ee, ee.t
}

// SetMax define the maximum number of reported events before listening
// for events is terminated automatically.  If m is 0 (or lower)
// listening doesn't stop automatically.
func (fx *Testing) SetMax(m int) *Events {
	switch {
	case m <= 0:
		if fx.ee.reported != nil {
			fx.ee.reported = nil
		}
		fx.autoTerminate = false
	default:
		fx.ee.reported = decrement(fx)
		fx.autoTerminate = true
	}
	fx.Max = m
	return fx.ee
}

const TestPanic = "test: can't call event triggering operation " +
	"in listener callback"

// waitForSynced waits on associated Events.Synced channel if not
// already waiting.  If already waiting the wait-stack is increased
// by given err and waitForSynced returns; leaving it to the currently
// waiting waitForSynced call to wait for this synchronization as well.
func (fx *Testing) waitForSynced(err string) {
	if fx.pushWaiting(err) { // return if already waiting
		return
	}
	tmr := time.NewTimer(fx.Timeout)
	for err := fx.popWaiting(); err != ""; err = fx.popWaiting() {
		select {
		case <-fx.ee.Synced:
			tmr.Reset(fx.Timeout)
		case <-tmr.C:
			fx.t.Fatalf(err)
		}
	}
	tmr.Stop()
}

// pushWaiting adds given string onto the wait-stack and returns true if
// if we are already waiting otherwise false and waiting is started.
func (fx *Testing) pushWaiting(err string) bool {
	fx.mutex.Lock()
	defer fx.mutex.Unlock()
	fx.waitStack = append(fx.waitStack, err)
	if fx.waiting {
		return true
	}
	fx.waiting = true
	return false
}

// popWaiting pops the first entry from the wait-stack and returns its
// error string unless the wait-stack is empty in which case the empty
// string is returned and we stop *waiting*.
func (fx *Testing) popWaiting() string {
	fx.mutex.Lock()
	defer fx.mutex.Unlock()
	if len(fx.waitStack) == 0 {
		fx.waiting = false
		return ""
	}
	err := fx.waitStack[0]
	fx.waitStack = fx.waitStack[1:]
	return err
}

// FireResize posts a resize event and returns after this event
// has been processed.  Is associated Events instance not listening
// it is started before the event is fired.
func (fx *Testing) FireResize(lines int) *Events {
	fx.t.Helper()
	if !fx.ee.IsListening() {
		fx.listen()
	}
	w, _ := fx.lib.Size()
	fx.lib.SetSize(w, lines)
	err := fx.lib.PostEvent(tcell.NewEventResize(w, lines))
	if err != nil {
		fx.t.Fatal(err)
	}
	fx.waitForSynced("test: set number of lines: sync timed out")
	fx.checkTermination()
	return fx.ee
}

// FireRune posts given run-key-press event and returns after this
// event has been processed.  Note modifier keys are ignored for
// rune-triggered key-events.  Are wrapped Events not polling it is
// started (ee.Listen()).
func (fx *Testing) FireRune(r rune) *Events {
	if !fx.ee.IsListening() {
		fx.listen()
	}
	fx.lib.InjectKey(tcell.KeyRune, r, tcell.ModNone)
	fx.waitForSynced("test: fire rune: sync timed out")
	fx.checkTermination()
	return fx.ee
}

// FireKey posts given special-key event and returns after this
// event has been processed.  Are wrapped Events not polling it is
// started (ee.Listen()).
func (fx *Testing) FireKey(k tcell.Key, m ...tcell.ModMask) *Events {
	fx.t.Helper()
	if !fx.ee.IsListening() {
		fx.listen()
	}
	if len(m) == 0 {
		fx.lib.InjectKey(k, 0, tcell.ModNone)
	} else {
		fx.lib.InjectKey(k, 0, m[0])
	}
	fx.waitForSynced("test: fire key: sync timed out")
	fx.checkTermination()
	return fx.ee
}

// listen posts the initial resize event and starts listening for events
// in a new go-routine.  listen returns after the initial resize has
// completed.
func (fx *Testing) listen() *Events {
	fx.t.Helper()
	err := fx.lib.PostEvent(tcell.NewEventResize(fx.lib.Size()))
	if err != nil {
		fx.t.Fatalf("test: listen: post resize: %v", err)
	}
	go fx.ee.listen()
	fx.waitForSynced("test: listen: sync timed out")
	fx.checkTermination()
	return fx.ee
}

func (fx *Testing) checkTermination() {
	if !fx.autoTerminate {
		return
	}
	if fx.Max <= 0 {
		// the last reported event might was a quit event,
		if fx.ee.IsListening() { // i.e. we stopped already listening
			fx.ee.QuitListening()
			fx.waitForSynced("quit listening: sync timed out")
		}
	}
}

func (fx *Testing) beforeFinalize() {
	fx.LastScreen = fx.String()
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
func (ee *Testing) String() string {
	b, w, h := ee.lib.GetContents()
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
