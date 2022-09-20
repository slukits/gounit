// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"log"
	"strings"

	"github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type fxInit struct {
	t *gounit.T

	// fatal is provided to the view to report fatal errors; it defaults
	// to log.Fatal
	fatal func(...interface{})

	// updateMessage holds the updater for the message bar which were
	// received through the Message implementation.
	updateMessage func(string)

	// updateStatus holds the updater for the status bar which were
	// received through the Status implementation.
	updateStatus func(Statuser)

	bttOneReported, bttTwoReported, bttThreeReported bool

	updButton func(Buttoner)

	// updateReporting NOTE if a reported line is not flagged with a
	// *focusable* LineMask the line is not selectable, i.e. is not
	// focused through key-events and not reported if clicked on it.
	updateReporting func(Reporter)

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
	fx.updateMessage = upd
	return fxMsg
}

func (fx *fxInit) Status(upd func(Statuser)) {
	fx.updateStatus = upd
}

func (fx *fxInit) Reporting(ru func(Reporter)) Reporter {
	fx.updateReporting = ru

	if fx.listenReporting == nil {
		fx.listenReporting = func(idx int) {
			fx.reportedLine = idx
		}
	}

	return &reporterFX{content: fxReporting, listener: fx.listenReporting}
}

func (fx *fxInit) Buttons(upd func(Buttoner)) Buttoner {
	fx.updButton = upd
	return &buttonerFX{
		newBB: []ButtonDef{
			{fxBtt1, fxRnBtt1}, {fxBtt2, fxRnBtt2}, {fxBtt3, fxRnBtt3}},
		listener: func(label string) {
			switch label {
			case fxBtt1, fxBtt1Upd:
				fx.bttOneReported = true
			case fxBtt2:
				fx.bttTwoReported = true
			case fxBtt3:
				fx.bttThreeReported = true
			}
		},
	}
}

// viewFX encapsulates the white-box aspects of this package's tests and
// augments the embedded view-instance with convenience methods for
// testing.
type viewFX struct {
	*view
	*fxInit
	Report *report
}

func newFX(t *gounit.T) *viewFX {
	fx := viewFX{}
	fx.fxInit = &fxInit{t: t}
	fx.view = New(fx.fxInit)
	fx.Report = fx.CC[1].(*report)
	return &fx
}

type fixtureSetter interface{ Set(*gounit.T, interface{}) }

func fx(t *gounit.T, fs fixtureSetter) (
	*lines.Events, *lines.Testing, *viewFX,
) {
	fx := newFX(t)
	ee, tt := lines.Test(t.GoT(), fx)
	ee.Listen()
	fs.Set(t, ee.QuitListening)
	return ee, tt, fx
}

func (fx *viewFX) ClickButton(tt *lines.Testing, label string) {
	if len(fx.CC) < 4 {
		fx.t.Fatal("gounit: view: fixture: expected 4 ui components")
	}
	bb, ok := fx.CC[3].(*buttonBar)
	if !ok {
		fx.t.Fatal("gounit: view: fixture: " +
			"expected forth component to be a button bar")
	}
	for _, b := range bb.bb {
		if b.label != label {
			continue
		}
		tt.FireComponentClick(b, 0, 0)
		return
	}
	fx.t.Fatalf("gounit: view: fixture: no button labeled %q", label)
}

func (fx *viewFX) twoPointFiveTimesReportedLines() (int, string) {
	len, lastLine := 0, ""
	fx.updateReporting(&reporterFX{f: func(
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
