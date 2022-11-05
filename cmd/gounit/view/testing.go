// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"log"
	"strings"
	"time"

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
	*lines.Fixture
	*view

	// ReportCmp provides a view's reporting component.
	ReportCmp *report

	// MsgCmp provides a view's message bar component.
	MsgCmp *messageBar

	// Cmp provides the embedded Component of a view's root component.
	Cmp *lines.Component

	// UpdateMessage holds the updater for the message bar which was
	// received through the Message implementation of the initer.
	UpdateMessage func(string)

	// UpdateStatus holds the updater for the status bar which was
	// received through the Status implementation of the initer.
	UpdateStatus func(Statuser)

	// UpdateButtons holds the updater for the button bar which was
	// received through the Buttons implementation of the initer.
	UpdateButtons func(Buttoner)

	// UpdateReporting holds the updater for the reporting component
	// which was received through the Reporting implementation of the
	// initer. NOTE if a reported line is not flagged with a *focusable*
	// LineMask the line is not selectable, i.e. is not focused through
	// key-events and not reported if clicked on it.
	UpdateReporting func(Reporter)

	ReportedButton string

	ReportedLine int
}

// Fixture creates a new view testing fixture instance embedding created
// view and lines.Testing-instance and augmenting their methods with
// some convenience methods for testing.
func Fixture(t *gounit.T, timeout time.Duration, i Initer) *Testing {
	tt := &Testing{T: t}
	var vw *view
	if i == nil {
		vw = New(&fxInit{t: t, tt: tt})
	} else {
		vw = New(&fxInit{t: t, tt: tt, initer: i})
	}
	tt.view = vw
	tt.Fixture = lines.TermFixture(t.GoT(), timeout, vw)
	tt.ReportCmp = tt.getReporting()
	tt.MsgCmp = tt.getMessageBar()
	tt.Cmp = &vw.Component
	return tt
}

// Fixture creates a new view testing fixture instance embedding created
// view and lines.Testing-instance and augmenting their methods with
// some convenience methods for testing.
func FixtureFor(
	t *gounit.T, timeout time.Duration, vw lines.Componenter,
) *Testing {
	_vw, ok := vw.(*view)
	if !ok {
		t.Fatalf("view: fixture for: expect componenter of type view "+
			"got: %T", vw)
	}
	return &Testing{T: t,
		Fixture:         lines.TermFixture(t.GoT(), timeout, vw),
		view:            _vw,
		UpdateMessage:   _vw.updateMessageBar,
		UpdateReporting: _vw.updateLines,
		UpdateStatus:    _vw.updateStatusBar,
		UpdateButtons:   _vw.updateButtons,
	}
}

// ClickButton clicks the button in the button-bar with given label.
// ClickButton does not return before subsequent view-changes triggered
// by requested button click are processed.
func (t *Testing) ClickButton(label string) {
	t.T.GoT().Helper()
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

// MessageBarCells returns the test-screen portion of the message bar.
func (t *Testing) MessageBarCells() lines.CellsScreen {
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
	return t.CellsOf(mb)
}

// ReportCells returns the test-screen portion of the reporting component.
func (t *Testing) ReportCells() lines.CellsScreen {
	rp := t.getReporting()
	if rp == nil {
		return nil
	}
	return t.CellsOf(rp)
}

// StatusBarCells returns the test-screen portion of the status bar.
func (t *Testing) StatusBarCells() lines.CellsScreen {
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
	return t.CellsOf(sb)
}

// ButtonBarCells returns the test-screen portion of the button bar.
func (t *Testing) ButtonBarCells() lines.CellsScreen {
	bb := t.getButtonBar()
	if bb == nil {
		return nil
	}
	return t.CellsOf(bb)
}

// TwoPointFiveTimesReportedLines fills the reporting area with 2.5
// times the displayed lines.  Handy for scrolling tests.
func (fx *Testing) TwoPointFiveTimesReportedLines() (int, string) {
	len, lastLine := 0, ""
	fx.UpdateReporting(&reporterFX{f: func(
		r lines.Componenter, f func(uint, string),
	) {
		n := 2*r.(*report).Dim().Height() + r.(*report).Dim().Height()/2
		for i := uint(0); i < uint(n); i++ {
			switch {
			case i < 10:
				lastLine = fmt.Sprintf("line 00%d", i)
				f(i, lastLine)
			case i < 100:
				lastLine = fmt.Sprintf("line 0%d", i)
				f(i, lastLine)
			default:
				lastLine = fmt.Sprintf("line %d", i)
				f(i, lastLine)
			}
		}
		len = r.(*report).Len()
	}})
	return len, lastLine
}

const notEnough = "gounit: view: fixture: not enough ui components"

func (t *Testing) getMessageBar() *messageBar {
	if len(t.CC) < 1 {
		t.T.Fatal(notEnough)
		return nil
	}
	rp, ok := t.CC[0].(*messageBar)
	if !ok {
		t.T.Fatal("gounit: view: fixture: " +
			"expected first component to be a message bar")
		return nil
	}
	return rp
}

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

func (t *Testing) defaultButtonListener(label string) {
	t.ReportedButton = label
}

func (t *Testing) defaultReportListener(idx int) {
	t.ReportedLine = idx
}

type fxInit struct {
	t      *gounit.T
	initer Initer
	tt     *Testing

	// fatal is provided to the view to report fatal errors; it defaults
	// to log.Fatal
	fatal func(...interface{})

	bttOneReported, bttTwoReported, bttThreeReported bool

	listenReporting func(int)

	reportedLine int
}

const (
	fxMsg       = "init fixture message"
	fxStatus    = "pkgs/suites: 0/0; tests: 0/0"
	fxReporting = "init fixture reporting"
	fxBtt1      = "first"
	fxBtt1Upd   = "hurz"
	fxBtt2      = "second"
	fxBtt3      = "third"
	fxRnBtt1    = '1'
	fxRnBtt2    = '2'
	fxRnBtt3    = '3'
)

// Fatal provides the function for fatal view-errors and defaults to
// log.Fatal.
func (fx *fxInit) Fatal() func(...interface{}) {
	if fx.fatal == nil {
		return log.Fatal
	}
	return fx.fatal
}

func (fx *fxInit) Message(upd func(string)) string {
	if fx.tt != nil {
		fx.tt.UpdateMessage = upd
	}
	if fx.initer != nil {
		return fx.initer.Message(upd)
	}
	return fxMsg
}

func (fx *fxInit) Status(upd func(Statuser)) {
	if fx.tt != nil {
		fx.tt.UpdateStatus = upd
	}
	if fx.initer != nil {
		fx.initer.Status(upd)
	}
}

func (fx *fxInit) Reporting(upd func(Reporter)) Reporter {
	if fx.tt != nil {
		fx.tt.UpdateReporting = upd
	}
	if fx.initer != nil {
		return fx.initer.Reporting(upd)
	}

	if fx.listenReporting == nil {
		if fx.tt != nil {
			fx.listenReporting = fx.tt.defaultReportListener
		} else {
			fx.listenReporting = func(idx int) {
				fx.reportedLine = idx
			}
		}
	}

	return &reporterFX{content: fxReporting, listener: fx.listenReporting}
}

func (fx *fxInit) Buttons(upd func(Buttoner)) Buttoner {
	if fx.tt != nil {
		fx.tt.UpdateButtons = upd
	}
	if fx.initer != nil {
		return fx.initer.Buttons(upd)
	}
	if fx.tt == nil {
		return &buttonerFX{
			newBB: []ButtonDef{
				{fxBtt1, fxRnBtt1}, {fxBtt2, fxRnBtt2}, {fxBtt3, fxRnBtt3}},
			listener: func(label string) {},
		}
	}
	return &buttonerFX{
		newBB: []ButtonDef{
			{fxBtt1, fxRnBtt1}, {fxBtt2, fxRnBtt2}, {fxBtt3, fxRnBtt3}},
		listener: fx.tt.defaultButtonListener,
	}
}

// reporterFX structure implements the Reporter interface to update the
// reporting component in tests.
type reporterFX struct {
	content  string
	ll       []string
	mm       map[uint]LineMask
	f        func(lines.Componenter, func(uint, string))
	flags    RprtMask
	listener func(int)
}

func (l *reporterFX) Flags() RprtMask { return l.flags }

func (l *reporterFX) LineMask(idx uint) LineMask {
	if l.mm == nil {
		return 0
	}
	return l.mm[idx]
}

func (l *reporterFX) Listener() func(int) { return l.listener }

func (l *reporterFX) For(r lines.Componenter, cb func(uint, string)) {
	if l.f != nil {
		l.f(r, cb)
		return
	}
	if len(l.ll) == 0 && l.content != "" {
		l.ll = strings.Split(l.content, "\n")
	}
	if len(l.ll) == 0 {
		return
	}
	for idx, l := range l.ll {
		if l == "" {
			continue
		}
		cb(uint(idx), l)
	}
}

type buttonerFX struct {
	replace  bool
	listener func(string)
	newBB    []ButtonDef
	updBB    map[string]ButtonDef
	err      func(error)
}

func (b buttonerFX) Replace() bool { return b.replace }

func (b *buttonerFX) ForUpdate(cb func(string, ButtonDef) error) {
	for label, def := range b.updBB {
		if err := cb(label, def); err != nil {
			if b.err == nil {
				panic(err)
			}
			b.err(err)
		}
	}
}

func (b *buttonerFX) ForNew(cb func(ButtonDef) error) {
	for _, def := range b.newBB {
		if err := cb(def); err != nil {
			if b.err == nil {
				panic(err)
			}
			b.err(err)
		}
	}
}

func (b *buttonerFX) Listener() ButtonLst {
	return b.listener
}
