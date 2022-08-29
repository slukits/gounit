// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type fxInit struct {
	t *gounit.T

	// update *bar are holding the updaters for message- and statusbar
	// which were received through the Status and Message implementations.
	updateMessageBar, updateStatusbar func(string)

	bttOneReported, bttTwoReported, bttThreeReported bool

	// updBtt* are the button updater received through the
	// implementation of ForButton
	updBtt1, updBtt2, updBtt3 func(ButtonDef) error

	// mainListener holds the lines listener updater received through
	// Main implementation
	mainListener LLUpdater

	// mainLines holds the lines updater received through Main
	// implementation
	mainLines LinesUpdater
}

const (
	fxMsg    = "init fixture message"
	fxStatus = "init fixture status"
	fxMain   = "init fixture main"
	fxBtt1   = "first"
	fxBtt2   = "second"
	fxBtt3   = ""
	fxRnBtt1 = '1'
	fxRnBtt2 = '2'
	fxRnBtt3 = '3'
)

func (fx *fxInit) Message(upd func(string)) string {
	fx.updateMessageBar = upd
	return fxMsg
}

func (fx *fxInit) Status(upd func(string)) string {
	fx.updateStatusbar = upd
	return fxStatus
}

func (fx *fxInit) Main(llu LLUpdater, lu LinesUpdater) string {
	fx.mainListener = llu
	fx.mainLines = lu
	return fxMain
}

func (fx *fxInit) ForButton(cb func(ButtonDef, ButtonUpdater) error) {
	cb(ButtonDef{
		Label:    fxBtt1,
		Rune:     fxRnBtt1,
		Listener: func(_ string) { fx.bttOneReported = true },
	}, func(update func(ButtonDef) error) { fx.updBtt1 = update })
	// button button two is reported but there will be no callback since
	// not listener is given
	cb(ButtonDef{Label: fxBtt2, Rune: fxRnBtt2},
		func(update func(ButtonDef) error) { fx.updBtt2 = update })
	// button three is not reported because its not part of the layout
	// with an zero-string label
	cb(ButtonDef{Label: fxBtt3, Rune: fxRnBtt3},
		func(update func(ButtonDef) error) { fx.updBtt3 = update })
}

// viewFX encapsulates the white-box aspects of this package's tests.
type viewFX struct {
	*view
	*fxInit
}

func newFX(t *gounit.T) *viewFX {
	fx := viewFX{}
	fx.fxInit = &fxInit{t: t}
	fx.view = New(fx.fxInit)
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
