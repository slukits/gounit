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

// ButtonUpdater is a callback function provided to the ForButton
// callback initializing a button.  To this callback a view instance
// provides an update function which can be used to update a button
// definition.  The later fails if the new button definition provides an
// ambiguous label.
type ButtonUpdater func(update func(ButtonDef) error)

// ButtonUpd is a callback updating, adding or removing a button.  The
// button with given label is removed if the button definition's label
// is zero.  A new button is added if given label is zero but not the
// button definition's.  The button is updated if given label and button
// definition's label is not zero.  Updating/adding fails if the button
// definition's label or rune is ambiguous.
type ButtonUpd func(label string, _ ButtonDef) error

// ButtonLst is informed about a button selection.
type ButtonLst func(label string)

type buttonBar struct {
	lines.Component
	lines.Chainer
	bb       []*button
	listener func(string)
}

func (sb *buttonBar) OnInit(_ *lines.Env) {
	sb.Dim().SetHeight(1)
}

func (sb *buttonBar) OnUpdate(e *lines.Env) {
	upd := e.Evt.(*lines.UpdateEvent).Data.(*buttonUpdate)
	switch {
	case upd.label == "":
		sb.append(upd.def)
		upd.setRune(upd.def.Rune, sb.button(upd.def.Label))
	case upd.label != "" && upd.def.Label == "":
		sb.delete(upd.label)
		upd.setRune(upd.def.Rune, nil)
	default:
		sb.update(e.EE, upd)
	}
}

func (sb *buttonBar) append(bd ButtonDef) {
	sb.bb = append(sb.bb, &button{
		label:    bd.Label,
		rn:       bd.Rune,
		listener: sb.listener,
	})
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

func (sb *buttonBar) update(ee *lines.Events, upd *buttonUpdate) {
	for _, b := range sb.bb {
		if b.label != upd.label {
			continue
		}
		ee.Update(b, upd, nil)
	}
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

func (sb *buttonBar) ForChained(
	cb func(lines.Componenter) (stop bool),
) {
	for _, b := range sb.bb {
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
	fmt.Fprint(e, b.uiLabel())
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
	bd := e.Evt.(*lines.UpdateEvent).Data.(*buttonUpdate)
	b.label = bd.def.Label
	b.rn = bd.def.Rune
	fmt.Fprint(e, b.uiLabel())
	bd.setRune(b.rn, b)
}
