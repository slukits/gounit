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

	// update *bar are holding the updaters for message- and statusbar
	// which were received through the Status and Message implementations.
	updateMessage, updateStatus func(string)

	bttOneReported, bttTwoReported, bttThreeReported bool

	updButton ButtonUpd

	updateReporting ReportingUpd

	listenReporting ReportingLst

	reportedLine int
}

const (
	fxMsg       = "init fixture message"
	fxStatus    = "init fixture status"
	fxReporting = "init fixture reporting"
	fxBtt1      = "first"
	fxBtt1Upd   = "hurz"
	fxBtt2      = "second"
	fxBtt3      = ""
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

func (fx *fxInit) Status(upd func(string)) string {
	fx.updateStatus = upd
	return fxStatus
}

func (fx *fxInit) Reporting(ru ReportingUpd) (string, ReportingLst) {
	fx.updateReporting = ru

	if fx.listenReporting == nil {
		fx.listenReporting = func(idx int) {
			fx.reportedLine = idx
		}
	}

	return fxReporting, fx.listenReporting
}

func (fx *fxInit) Buttons(
	bu ButtonUpd, For func(ButtonDef) error,
) ButtonLst {

	fx.updButton = bu
	df := []ButtonDef{
		{fxBtt1, fxRnBtt1}, {fxBtt2, fxRnBtt2}, {fxBtt3, fxRnBtt3}}
	for _, d := range df {
		if err := For(d); err != nil {
			panic(err)
		}
	}
	return func(label string) {
		switch label {
		case fxBtt1, fxBtt1Upd:
			fx.bttOneReported = true
		case fxBtt2:
			fx.bttTwoReported = true
		case fxBtt3:
			fx.bttThreeReported = true
		}
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
	fx.updateReporting(&linerFX{f: func(
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

type linerFX struct {
	content  string
	f        func(lines.Componenter, func(uint, string))
	clearing bool
}

func (l *linerFX) Clearing() bool { return l.clearing }

func (l *linerFX) Mask(idx uint) LineMask {
	return 0
}

func (l *linerFX) For(r lines.Componenter, cb func(uint, string)) {
	if l.f != nil {
		l.f(r, cb)
		return
	}
	if l.content == "" {
		return
	}
	ll := strings.Split(l.content, "\n")
	for idx, l := range ll {
		if l == "" {
			continue
		}
		cb(uint(idx), l)
	}
}
