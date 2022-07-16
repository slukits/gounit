// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
)

// Screen provides features to write to a line-based terminal
// user-interface.
type Screen struct {
	lib    tcell.Screen
	ll     *Lines
	errScr *ErrScr
	min    int
}

// Len returns the number of lines of wrapped terminal screen.  Note len
// of the simulation screen defaults to 25.
func (s *Screen) Len() int {
	_, h := s.lib.Size()
	return h
}

// LL returns the screens's currently focused lines-set.
func (s *Screen) LL() *Lines { return s.ll }

// SetMin defines the minimal expected number of screen lines.  An error
// is displayed and event reporting is suppressed as long as the
// screen-height is below min.
func (s *Screen) SetMin(m int) {
	s.min = m
	if s.Len() > s.min {
		return
	}
	s.minErr()
}

// ToSmall returns true if a set minimal number of screen lines is
// greater than the available screen lines.
func (s *Screen) ToSmall() bool { return s.Len() < s.min }

func (s *Screen) resize() (ok bool) {
	s.lib.Clear()
	s.ll.ensure(s.Len())
	ok = s.Len() > s.min
	if ok {
		if s.errScr != nil && s.errScr.Active {
			s.errScr.Active = false
		}
		return ok
	}
	s.minErr()
	return ok
}

// ErrScreenFmt is the displayed error message for the case that a set
// minimal number of lines is greater than the available screen height.
const ErrScreenFmt = "minimum screen-height: %d"

func (s *Screen) minErr() {
	if !s.ErrScreen().Active {
		s.ErrScreen().Active = true
	}

	if s.ErrScreen().String() != fmt.Sprintf(ErrScreenFmt, s.min) {
		s.ErrScreen().Set(fmt.Sprintf(ErrScreenFmt, s.min))
	}
}

func (s *Screen) ensureSynced(show bool) {
	sync := func() {
		if show {
			s.lib.Show()
		} else {
			s.lib.Sync()
		}
	}
	if s.errScr != nil {
		if s.errScr.Active {
			if s.errScr.isDirty {
				s.errScr.sync()
				sync()
			}
			return
		}
	}
	if !s.ll.isDirty() {
		return
	}
	s.ll.sync()
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
