// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"github.com/slukits/gounit"
	"github.com/slukits/lines"
)

// Testing augments a view-instance with functionality useful for
// testing but not meant for production.  A Testing-view instance may be
// initialized by
//
//	tt := view.Testing{t, view.New(i)}
//
// whereas t is an *gounit.T and i and view.Initier implementation.
type Testing struct {
	T *gounit.T
	*lines.Testing
	*view
}

func NewTesting(
	t *gounit.T, tt *lines.Testing, c lines.Componenter,
) *Testing {
	vw, ok := c.(*view)
	if !ok {
		t.Fatal("given component must be a view; got %T", c)
		return nil
	}
	return &Testing{T: t, Testing: tt, view: vw}
}

// ClickButton clicks the button in the button-bar with given label.
// ClickButton does not return before subsequent view-changes triggered
// by requested button click are processed.
func (t *Testing) ClickButton(label string) {
	bb := t.getButtonBar()
	if bb == nil {
		return
	}
	for _, b := range bb.bb {
		if b.label != label {
			continue
		}
		t.FireComponentClick(b, 0, 0)
		return
	}
	t.T.Fatalf("gounit: view: fixture: no button labeled %q", label)
}

// ClickReporting clicks on the line with given index of the view's
// reporting component.  ClickReporting does not return before
// subsequent view-changes triggered by requested reporting click are
// processed.
func (t *Testing) ClickReporting(idx int) {
	rp := t.getReporting()
	if rp == nil {
		return
	}
	t.FireComponentClick(rp, 0, idx)
}

// MessageBar returns the test-screen portion of the message bar.
func (t *Testing) MessageBar() lines.TestScreen {
	if len(t.CC) < 1 {
		t.T.Fatal("gounit: view: fixture: no ui components")
		return nil
	}
	mb, ok := t.CC[0].(*messageBar)
	if !ok {
		t.T.Fatal("gounit: view: fixture: " +
			"expected first component to be the message bar")
		return nil
	}
	return t.ScreenOf(mb)
}

// Reporting returns the test-screen portion of the reporting component.
func (t *Testing) Reporting() lines.TestScreen {
	rp := t.getReporting()
	if rp == nil {
		return nil
	}
	return t.ScreenOf(rp)
}

// StatusBar returns the test-screen portion of the status bar.
func (t *Testing) StatusBar() lines.TestScreen {
	if len(t.CC) < 3 {
		t.T.Fatal(notEnough)
		return nil
	}
	sb, ok := t.CC[2].(*statusBar)
	if !ok {
		t.T.Fatal("gounit: view: fixture: " +
			"expected third component to be the status bar")
		return nil
	}
	return t.ScreenOf(sb)
}

// ButtonBar returns the test-screen portion of the button bar.
func (t *Testing) ButtonBar() lines.TestScreen {
	bb := t.getButtonBar()
	if bb == nil {
		return nil
	}
	return t.ScreenOf(bb)
}

// Trim trims a given test-screen portion horizontally and vertically,
// i.e. is a short cut for ts.TrimHorizontal().TrimVertical().
func (t *Testing) Trim(ts lines.TestScreen) lines.TestScreen {
	return ts.TrimHorizontal().TrimVertical()
}

const notEnough = "gounit: view: fixture: not enough ui components"

func (t *Testing) getReporting() *report {
	if len(t.CC) < 2 {
		t.T.Fatal(notEnough)
		return nil
	}
	rp, ok := t.CC[1].(*report)
	if !ok {
		t.T.Fatal("gounit: view: fixture: " +
			"expected second component to be reporting")
		return nil
	}
	return rp
}

func (t *Testing) getButtonBar() *buttonBar {
	if len(t.CC) < 4 {
		t.T.Fatal(notEnough)
		return nil
	}
	bb, ok := t.CC[3].(*buttonBar)
	if !ok {
		t.T.Fatal("gounit: view: fixture: " +
			"expected forth component to be a button bar")
		return nil
	}
	return bb
}
