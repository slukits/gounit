// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/ints"
)

// View provides a line-based terminal user-interface.  A view instances
// are provided to event-listeners.  To avoid race conditions they are
// not meant to be created or kept around outside an event listener
// callback operation.
type View struct {
	lib       tcell.Screen
	ll        *Lines
	evtRunes  string
	evtKeys   ints.Set
	isPolling bool
	errScr    *ErrScr
	min       int

	// Synced provides a message each time it is guaranteed that the
	// potential effect of an event is on the screen.
	Synced chan bool
}

// Len returns the number of lines of wrapped terminal screen.  Note len
// of the simulation screen defaults to 25.
func (v *View) Len() int {
	_, h := v.lib.Size()
	return h
}

// LL returns the receiving view's lines-set.
func (v *View) LL() *Lines { return v.ll }

// SetMin defines the minimal expected number of screen lines.  An error
// is displayed and event reporting is suppressed as long as the
// screen-height is below min.
func (v *View) SetMin(m int) {
	v.min = m
	if v.Len() > v.min {
		return
	}
	v.minErr()
}

// ToSmall returns true if a set minimal number of lines is greater than
// the available screen height.
func (v *View) ToSmall() bool { return v.Len() < v.min }

func (v *View) resize() (ok bool) {
	v.lib.Clear()
	v.ll.ensure(v.Len())
	ok = v.Len() > v.min
	if ok {
		if v.errScr != nil && v.errScr.Active {
			v.errScr.Active = false
		}
		return ok
	}
	v.minErr()
	return ok
}

// ErrScreenFmt is the displayed error message for the case that a set
// minimal number of lines is greater than the available screen height.
const ErrScreenFmt = "minimum screen-height: %d"

func (v *View) minErr() {
	if !v.ErrScreen().Active {
		v.ErrScreen().Active = true
	}

	if v.ErrScreen().String() != fmt.Sprintf(ErrScreenFmt, v.min) {
		v.ErrScreen().Set(fmt.Sprintf(ErrScreenFmt, v.min))
	}
}

func (v *View) ensureSynced(show bool) {
	sync := func() {
		if show {
			v.lib.Show()
		} else {
			v.lib.Sync()
		}
	}
	if v.errScr != nil {
		if v.errScr.Active {
			if v.errScr.isDirty {
				v.errScr.sync()
				sync()
			}
			return
		}
	}
	if !v.ll.isDirty() {
		return
	}
	v.ll.sync()
	sync()
}

// screenFactory is used to create new tcell-screens for production or
// for simulation.  export_test.go makes it possible to replace this
// screen factory with a screen-factory mocking up tcell's screen
// creation errors so they can be tested.
var screenFactory screenFactoryer = &defaultFactory{}

type defaultFactory struct{}

func (f *defaultFactory) NewScreen() (tcell.Screen, error) {
	return tcell.NewScreen()
}

func (f *defaultFactory) NewSimulationScreen(
	s string,
) tcell.SimulationScreen {
	return tcell.NewSimulationScreen(s)
}

type screenFactoryer interface {
	NewScreen() (tcell.Screen, error)
	NewSimulationScreen(string) tcell.SimulationScreen
}
