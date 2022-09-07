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

type ReportingLst func(idx int)

type ReportingUpd func(Liner)

type report struct {
	lines.Component
	dflt     string
	listener func(int)
}

func (m *report) OnInit(e *lines.Env) {
	m.FF.Add(lines.Scrollable)
	fmt.Fprint(e, m.dflt)
}

func (m *report) OnClick(_ *lines.Env, _, y int) {
	if m.listener == nil {
		return
	}
	m.listener(y)
}

func (m *report) OnUpdate(e *lines.Env) {
	data := e.Evt.(*lines.UpdateEvent).Data
	switch dt := data.(type) {
	case Liner:
		ii := &ints.Set{}
		dt.For(m, func(idx uint, content string) {
			ii.Add(int(idx))
			fmt.Fprint(m.wrt(dt, idx, e), content)
		})
		if !dt.Clearing() {
			return
		}
		for i := 0; i < m.Len(); i++ {
			if ii.Has(i) {
				continue
			}
			m.Reset(i)
		}
	case func(int):
		m.listener = dt
	}
}

func (m *report) wrt(l Liner, idx uint, e *lines.Env) io.Writer {
	switch l.Mask(idx) {
	case Failed:
		return e.BG(tcell.ColorRed).
			FG(tcell.ColorWhite).
			LL(int(idx))
	default:
		return e.LL(int(idx))
	}
}

// OnContext scrolls given reporting component down.  If at bottom it is
// scrolled to the top.
func (r *report) OnContext(e *lines.Env, x, y int) {
	if r.Scroll.IsAtBottom() {
		r.Scroll.ToTop()
		return
	}
	r.Scroll.Down()
}
