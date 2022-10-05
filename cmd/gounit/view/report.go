// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/lines"
)

// RprtMask types flags for Reporter-implementations.
type RprtMask uint

const (

	// RpClearing indicates that all lines of a view's reporting
	// component which are not updated by an Reporter-implementation are
	// cleared.
	RpClearing RprtMask = 1 << iota

	// RpNoFlags is the return value of an Reporter.Flags implementation
	// where no flags are set.
	RpNoFlags = 0
)

// LineMask  values are used to describe to a reporting component how a
// particular line should be displayed and if it is focusable or not.
type LineMask uint64

const (

	// Failed sets error formattings for a reporting component's line
	// like a red background and a white foreground.
	Failed LineMask = 1 << iota

	// Passed sets a lines as the "green bar", i.e. a green background
	// and a black foreground.
	Passed

	// PackageLine classifies a reported line as package-line.  Note
	// only package or suite lines are selectable.
	PackageLine

	// PackageFoldedLine reports a folded package.
	PackageFoldedLine

	// TestLine classifies a reported line as go-test-line.
	TestLine

	// OutputLine classifies a reported line as a test-output.
	OutputLine

	// GoTestsLine classifies a reported line as the go-tests headline
	// in a package report having go tests and test-suites.
	GoTestsLine

	// GoTestsFoldedLine classifies the headline of the go tests from a
	// reported testing package which has go-tests and test-suite.
	GoTestsFoldedLine

	// GoSuiteLine classifies a reported line as go test with sub-tests.
	// Note only package, go-tests, go-suite or suite lines are
	// selectable.
	GoSuiteLine

	// GoSuiteFolded classifies a go-test line with folded sub-tests.
	GoSuiteFoldedLine

	// SuiteLine classifies a reported line as test-suite.
	SuiteLine

	// SuiteFoldedLine classifies a reported line as a test-suite whose
	// suite-tests are folded.
	SuiteFoldedLine

	// SuiteTestLine classifies a reported line as suit-test-line.  Note
	// a suit-test-line is not selectable.
	SuiteTestLine

	// ZeroLineMode indicates no other than default formattings for a
	// line of a reporting component.
	ZeroLineMod LineMask = 0
)

const focusable LineMask = PackageLine | PackageFoldedLine |
	GoTestsLine | GoTestsFoldedLine | GoSuiteLine | GoSuiteFoldedLine |
	SuiteLine | SuiteFoldedLine

// A Reporter implementation provides line-updates for the gounit view's
// reporting area.
type Reporter interface {

	// Flags returns an optional combination of flags controlling how a
	// given Reporter implementation is processed.  See Rp*-constants.
	Flags() RprtMask

	// For is provided with the reporting component instance and a
	// callback function which must be called for each line which should
	// be updated.  If Clearing is ture all other lines of the reporting
	// component are reset to zero.  For each updated line Mask is
	// called for optional formatting information.
	For(_ lines.Componenter, line func(idx uint, content string))

	// LineMask may provide for an updated line additional formatting
	// information like "Failed" or "Passed" which accordingly adapts
	// the formatting of the line with given index.
	LineMask(idx uint) LineMask

	// Listener implementation of a Reporter provides a callback
	// function which is informed about line selections by the user
	// providing the selected line's id.
	Listener() func(idx int)
}

type report struct {
	lines.Component
	rr       []Reporter
	listener func(int)
}

func (m *report) OnInit(e *lines.Env) {
	m.FF.Add(lines.Scrollable | lines.LinesSelectable)
	m.rr[0].For(m, func(idx uint, content string) {
		fmt.Fprint(e.LL(int(idx)), content)
	})
	m.listener = m.rr[0].Listener()
}

// OnClick reports clicked line to given report-listener iff the line is
// focusable.
func (r *report) OnClick(_ *lines.Env, _, y int) {
	idx := r.Scroll.CoordinateToIndex(y)
	if r.listener == nil || y >= r.Len() || !r.LL(idx).IsFocusable() {
		return
	}
	r.listener(idx)
}

func (r *report) OnUpdate(e *lines.Env) {
	r.Focus.Reset(true)
	upd, ok := e.Evt.(*lines.UpdateEvent).Data.(Reporter)
	if !ok {
		return
	}
	r.Reset(lines.All, lines.NotFocusable)
	r.listener = upd.Listener()
	upd.For(r, func(idx uint, content string) {
		lm := upd.LineMask(idx)
		if lm&Failed > 0 {
			r.reportFailed(idx, lm, e, content)
			return
		}
		if lm&focusable == 0 {
			fmt.Fprint(e.LL(int(idx), lines.NotFocusable), content)
			return
		}
		fmt.Fprint(e.LL(int(idx)), content)
	})
}

func (r *report) reportFailed(
	idx uint, lm LineMask, e *lines.Env, content string,
) {
	sr := lines.SR{Style: tcell.StyleDefault.Background(tcell.ColorRed).
		Foreground(tcell.ColorWhite)}
	for _, r := range content {
		if r != ' ' {
			break
		}
		sr.IncrementStart()
	}
	ff := lines.LineFlags(0)
	if lm&focusable == 0 {
		ff = lines.NotFocusable
	}
	spl := strings.Split(content, lines.LineFiller)
	sr.SetEnd(len(spl[0]))
	fmt.Fprint(e.LL(int(idx), ff), content)
	e.AddStyleRange(int(idx), sr)
}

// OnContext scrolls given reporting component down.  If at bottom it is
// scrolled to the top.
func (r *report) OnContext(e *lines.Env, x, y int) {
	r.scroll()
}

// OnRune scrolls given reporting component down iff given rune is the
// space rune.  If at bottom it is scrolled to the top.
func (r *report) OnRune(e *lines.Env, rn rune) {
	if rn != ' ' {
		return
	}
	r.scroll()
}

func (r *report) scroll() {
	if r.Scroll.IsAtBottom() {
		r.Scroll.ToTop()
		return
	}
	r.Scroll.Down()
}

func (r *report) OnLineSelection(e *lines.Env, idx int) {
	r.listener(idx)
}
