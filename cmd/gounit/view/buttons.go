// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"

	"github.com/slukits/lines"
)

// ButtonDef defines a button's label and optional rune-event associated
// with defined button and a listener to which a button click or button
// rune-event is reported to.
type ButtonDef struct {
	Label    string
	Rune     rune
	Listener func(label string)
}

// ButtonUpdater is a callback function provided to the ForButton
// callback initializing a button.  To this callback a view instance
// provides an update function which can be used to update a button
// definition.  The later fails if the new button definition provides an
// ambiguous label.
type ButtonUpdater func(update func(ButtonDef) error)

type buttonBar struct {
	lines.Component
	lines.Chainer
	bb []*button
}

func (sb *buttonBar) OnInit(_ *lines.Env) {
	sb.Dim().SetHeight(1)
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
}

func (b *button) OnInit(env *lines.Env) {
	fmt.Fprintf(env, b.label)
}

func (b *button) OnClick(_ *lines.Env, _, _ int) {
	if b.listener == nil {
		return
	}
	b.listener(b.label)
}

func (b *button) OnUpdate(e *lines.Env) {
	bd := e.Evt.(*lines.UpdateEvent).Data.(*ButtonDef)
	b.label = bd.Label
	b.listener = bd.Listener
	fmt.Fprint(e, b.label)
}
