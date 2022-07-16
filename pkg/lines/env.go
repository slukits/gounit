// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import "github.com/gdamore/tcell/v2"

// Env is an environment provided to event listeners when they are
// called back.  It provides an encapsulated Screen's public API,
// information about the event which triggered the callback and features
// to communicate with the reporting operation.
//
// NOTE it is not save to provide an Env instance or one of its method's
// return values to an other go routine.  The former will likely result
// in a nil pointer panic the later in a race-condition.  It is save
// though to pass an environment's property values on to a goroutine.
// If you want to use concurrency use the Events.Update-method to obtain
// in the concurrent go routine an environment e.g.:
//
//     func myListener(e *lines.Env) {
//         // property ok; e or an e-method's return values NOT OK!
//         go myVeryHeavyOperation(e.EE)
//         e.Statusbar("executing very heavy operation").Busy()
//     }
//
//     func myVeryHeavyOperation(e.EE) {
//         var theAnswer int
//         // implementation of very heavy operation to find the answer
//         e.EE.Update(func(e *lines.Env)) {
//             e.Statusbar("done")
//             e.MessageBar().Styledf(
//                 Centered, "the answer is %d" theAnswer)
//         }
//     }
type Env struct {
	scr *Screen

	// EE is the Events instance providing given environment
	// instance.
	EE *Events

	// Evn is the tcell-event triggering the creation of a receiving
	// environment to report it back to a registered listener.
	Evn tcell.Event
}

// Len returns the number of lines of terminal screen.  Note len of the
// fixture-screen defaults to 25.
func (e *Env) Len() int { return e.scr.Len() }

// LL returns the screens's currently focused lines-set.
func (e *Env) LL() *Lines { return e.scr.LL() }

// SetMin defines the minimal expected number of screen lines.  An error
// is displayed and event reporting is suppressed as long as the
// screen-height is below min.
func (e *Env) SetMin(m int) { e.scr.SetMin(m) }

// ToSmall returns true if a set minimal number of screen lines is
// greater than the available screen lines.
func (e *Env) ToSmall() bool { return e.scr.ToSmall() }

// ErrScreen returns an overlaying (if activated) error-screen allowing
// to report errors without loosing the screen content at the time the
// error screen is requested
func (e *Env) ErrScreen() *ErrScr { return e.scr.ErrScreen() }

func (e *Env) reset() {
	e.scr = nil
	e.EE = nil
	e.Evn = nil
}
