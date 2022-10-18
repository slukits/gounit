// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"strings"

	"github.com/slukits/lines"
)

// ButtonDef defines a button's label and optional rune-event associated
// with defined button and a listener to which a button click or button
// rune-event is reported to.
type ButtonDef struct {
	Label string
	Rune  rune
}

// Buttoner implementation defines a view's button-bar.
type Buttoner interface {

	// Replace indicates if existing buttons should be replaced with
	// provided buttons.  In the later case ForUpdate is ignored.
	Replace() bool

	// Listener updates view's button bar's listener iff the second
	// value is ture.
	Listener() ButtonLst

	// ForUpdate calls back for each button which should be updated
	// whereas the button to update is identified by given label.
	ForUpdate(func(label string, _ ButtonDef) error)

	// ForNew calls back for each new button and appends it to the
	// button-bar.
	ForNew(func(ButtonDef) error)
}

// ButtonLst is informed about a button selection.
type ButtonLst func(label string)

type buttonBar struct {
	lines.Component
	lines.Chainer
	bb       []*button
	listener func(string)
}

func (bb *buttonBar) OnInit(_ *lines.Env) {
	bb.Dim().SetHeight(1)
}

func (bb *buttonBar) OnUpdate(e *lines.Env) {
	upd := e.Evt.(*lines.UpdateEvent).Data.(*buttonsUpdate)
	if upd.bb.Replace() {
		bb.init(upd)
		return
	}
	bb.update(e.Lines, upd)
}

func (bb *buttonBar) init(bbDef *buttonsUpdate) {
	bb.bb = []*button{}
	bb.setListener(bbDef.bb.Listener())
	bbDef.bb.ForNew(func(bd ButtonDef) error {
		bb.append(bd, bbDef.setRune)
		return nil
	})
}

func (bb *buttonBar) update(ee *lines.Lines, bbDef *buttonsUpdate) {
	bbDef.bb.ForNew(func(bd ButtonDef) error {
		bb.append(bd, bbDef.setRune)
		return nil
	})
	bbDef.bb.ForUpdate(func(label string, bd ButtonDef) error {
		if bd.Label == "" {
			bb.delete(label)
			bbDef.setRune(0, label, nil)
			return nil
		}
		b := bb.button(label)
		ee.Update(b, bd, nil)
		bbDef.setRune(bd.Rune, label, b)
		return nil
	})
}

func (bb *buttonBar) append(
	bd ButtonDef, setRune func(rune, string, *button),
) {
	bb.bb = append(bb.bb, &button{
		label:    bd.Label,
		rn:       bd.Rune,
		listener: bb.listener,
	})
	if bd.Rune != 0 {
		setRune(bd.Rune, bd.Label, bb.bb[len(bb.bb)-1])
	}
}

func (sb *buttonBar) delete(label string) {
	idx := -1
	for i, b := range sb.bb {
		if b.label != label {
			continue
		}
		idx = i
	}
	if idx == -1 {
		return
	}
	copy(sb.bb[idx:], sb.bb[idx+1:])
	sb.bb = sb.bb[:len(sb.bb)-1]
}

func (sb *buttonBar) setListener(lst func(string)) {
	for _, b := range sb.bb {
		b.listener = lst
	}
	sb.listener = lst
}

func (sb *buttonBar) button(label string) *button {
	for _, b := range sb.bb {
		if b.label != label {
			continue
		}
		return b
	}
	return nil
}

func (bb *buttonBar) ForChained(
	cb func(lines.Componenter) (stop bool),
) {
	for _, b := range bb.bb {
		if b.label == "" {
			continue
		}
		cb(b)
	}
}

type button struct {
	lines.Component
	label    string
	listener func(string)
	rn       rune
}

func (b *button) OnInit(e *lines.Env) {
	lbl := b.uiLabel()
	if strings.HasSuffix(lbl, "=on") {
		b.Dim().SetWidth(len(lbl) + 1)
	} else {
		b.Dim().SetWidth(len(lbl))
	}
	fmt.Fprint(e, lbl)
}

func (b *button) uiLabel() string {
	lbl := b.label
	if strings.ContainsRune(lbl, b.rn) {
		lbl = strings.Replace(
			lbl, string(b.rn), fmt.Sprintf("[%c]", b.rn), 1)
	}
	return lbl
}

func (b *button) OnClick(_ *lines.Env, _, _ int) {
	if b.listener == nil {
		return
	}
	b.listener(b.label)
}

func (b *button) OnUpdate(e *lines.Env) {
	bd := e.Evt.(*lines.UpdateEvent).Data.(ButtonDef)
	b.label = bd.Label
	b.rn = bd.Rune
	fmt.Fprint(e, b.uiLabel())
}
