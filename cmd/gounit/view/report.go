// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"strings"

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

// Focusable combines all line mode flags of selectable lines.
const Focusable LineMask = PackageLine | PackageFoldedLine |
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
	// be updated.  If Clearing is true all other lines of the reporting
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
	m.LL.Focus.Trimmed()
	m.rr[0].For(m, func(idx uint, content string) {
		fmt.Fprint(e.LL(int(idx)), content)
	})
	m.listener = m.rr[0].Listener()
}

// OnClick reports clicked line to given report-listener iff the line is
// focusable.
func (r *report) OnClick(_ *lines.Env, _, y int) {
	idx := r.Scroll.CoordinateToIndex(y)
	if r.listener == nil || y >= r.Len() || r.LL.By(idx).IsFlagged(lines.NotFocusable) {
		return
	}
	r.listener(idx)
}

func (r *report) OnUpdate(e *lines.Env, data interface{}) {
	r.LL.Focus.Reset()
	upd, ok := data.(Reporter)
	if !ok {
		return
	}
	r.Reset(lines.All, lines.NotFocusable)
	r.listener = upd.Listener()
	upd.For(r, func(idx uint, content string) {
		lm := upd.LineMask(idx)
		if lm&Focusable == 0 {
			idx, indent := int(idx), indent(content)
			if !r.LL.By(idx).IsFlagged(lines.NotFocusable) {
				r.LL.By(idx).Switch(lines.NotFocusable)
			}
			lines.Print(e.LL(idx).At(0), []rune(content)[:indent])
			w := e.LL(idx).At(indent)
			if lm&Failed != 0 {
				w = w.FG(lines.White).BG(lines.DarkRed)
			}
			lines.Print(w, []rune(content)[indent:])
			return
		}
		r.reportFocusable(int(idx), lm, e, content)
	})
}

func (r *report) reportFocusable(
	idx int, lm LineMask, e *lines.Env, content string,
) {
	if r.LL.By(idx).IsFlagged(lines.NotFocusable) {
		r.LL.By(idx).Switch(lines.NotFocusable)
	}
	indent := indent(content)
	cc := strings.Split(content, lines.Filler)
	lines.Print(e.LL(idx).At(0), []rune(cc[0][:indent]))
	w := e.LL(idx).At(indent).AA(lines.Underline)
	if lm&Failed != 0 {
		w = w.FG(lines.White).BG(lines.DarkRed)
	}
	lines.Print(w, []rune(cc[0][indent:]))
	if len(cc) == 1 {
		return
	}
	lines.Print(
		e.LL(idx).At(len([]rune(cc[0]))),
		[]rune(lines.Filler+strings.Join(cc[1:], lines.Filler)),
	)
}

func indent(content string) int {
	indent := 0
	for _, r := range content {
		if r != ' ' {
			break
		}
		indent++
	}
	return indent
}

// OnContext scrolls given reporting component down.  If at bottom it is
// scrolled to the top.
func (r *report) OnContext(e *lines.Env, x, y int) {
	r.scroll()
}

// OnRune scrolls given reporting component down iff given rune is the
// space rune.  If at bottom it is scrolled to the top.
func (r *report) OnRune(e *lines.Env, rn rune, mm lines.ModifierMask) {
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
