// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"io"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/ints"
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

func (m *report) OnClick(_ *lines.Env, _, y int) {
	if m.listener == nil {
		return
	}
	m.listener(y)
}

func (m *report) OnUpdate(e *lines.Env) {
	m.Focus.Reset(true)
	r, ok := e.Evt.(*lines.UpdateEvent).Data.(Reporter)
	if !ok {
		return
	}
	clearing := r.Flags()&RpClearing == RpClearing
	ii := &ints.Set{}
	r.For(m, func(idx uint, content string) {
		ii.Add(int(idx))
		fmt.Fprint(m.wrt(r, idx, e), content)
	})
	if !clearing {
		return
	}
	for i := 0; i < m.Len(); i++ {
		if ii.Has(i) {
			continue
		}
		m.Reset(i, lines.NotFocusable)
	}
	m.listener = r.Listener()
}

func (m *report) wrt(l Reporter, idx uint, e *lines.Env) io.Writer {
	lm := l.LineMask(idx)
	var lf lines.LineFlags
	if lm&focusable == 0 {
		lf = lines.NotFocusable
	}
	if lm&Failed > 0 {
		return e.BG(tcell.ColorRed).
			FG(tcell.ColorWhite).
			LL(int(idx), lf)
	}
	return e.LL(int(idx), lf)
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
