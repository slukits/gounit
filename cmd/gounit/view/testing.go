// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"github.com/slukits/gounit"
	"github.com/slukits/lines"
)

// Test augments a view-instance with functionality useful for
// testing but not meant for production.  A Test-view instance may be
// initialized by
//
//	tt := view.Test{t, view.New(i)}
//
// whereas t is an *gounit.T and i and view.Initier implementation.
type Test struct {
	t *gounit.T
	*view
}

func (t *Test) ClickButton(tt *lines.Testing, label string) {
	bb := t.getButtonBar()
	if bb == nil {
		return
	}
	for _, b := range bb.bb {
		if b.label != label {
			continue
		}
		tt.FireComponentClick(b, 0, 0)
		return
	}
	t.t.Fatalf("gounit: view: fixture: no button labeled %q", label)
}

func (t *Test) ClickReporting(tt *lines.Testing, idx int) {
	rp := t.getReporting()
	if rp == nil {
		return
	}
	tt.FireComponentClick(rp, 0, idx)
}

func (t *Test) MessageBar(tt *lines.Testing) lines.TestScreen {
	if len(t.CC) < 1 {
		t.t.Fatal("gounit: view: fixture: no ui components")
		return nil
	}
	mb, ok := t.CC[0].(*messageBar)
	if !ok {
		t.t.Fatal("gounit: view: fixture: " +
			"expected first component to be the message bar")
		return nil
	}
	return tt.ScreenOf(mb)
}

func (t *Test) Reporting(tt *lines.Testing) lines.TestScreen {
	rp := t.getReporting()
	if rp == nil {
		return nil
	}
	return tt.ScreenOf(rp)
}

func (t *Test) StatusBar(tt *lines.Testing) lines.TestScreen {
	if len(t.CC) < 3 {
		t.t.Fatal(notEnough)
		return nil
	}
	sb, ok := t.CC[2].(*statusBar)
	if !ok {
		t.t.Fatal("gounit: view: fixture: " +
			"expected third component to be the status bar")
		return nil
	}
	return tt.ScreenOf(sb)
}

func (t *Test) ButtonBar(tt *lines.Testing) lines.TestScreen {
	bb := t.getButtonBar()
	if bb == nil {
		return nil
	}
	return tt.ScreenOf(bb)
}

func (t *Test) Trim(ts lines.TestScreen) lines.TestScreen {
	return ts.TrimHorizontal().TrimVertical()
}

const notEnough = "gounit: view: fixture: not enough ui components"

func (t *Test) getReporting() *report {
	if len(t.CC) < 2 {
		t.t.Fatal(notEnough)
		return nil
	}
	rp, ok := t.CC[1].(*report)
	if !ok {
		t.t.Fatal("gounit: view: fixture: " +
			"expected second component to be reporting")
		return nil
	}
	return rp
}

func (t *Test) getButtonBar() *buttonBar {
	if len(t.CC) < 4 {
		t.t.Fatal(notEnough)
		return nil
	}
	bb, ok := t.CC[3].(*buttonBar)
	if !ok {
		t.t.Fatal("gounit: view: fixture: " +
			"expected forth component to be a button bar")
		return nil
	}
	return bb
}
