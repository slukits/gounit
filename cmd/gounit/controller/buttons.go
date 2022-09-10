// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import "github.com/slukits/gounit/cmd/gounit/view"

type buttons struct {
	viewUpd    func(interface{})
	lastReport view.Liner
	dflt       *buttoner
	more       *buttoner
	args       *buttoner
}

func newButtons(upd func(interface{}), lastReport view.Liner) *buttons {
	return &buttons{viewUpd: upd, lastReport: lastReport}
}

func (bb *buttons) defaultButtons() *buttoner {
	if bb.dflt == nil {
		bb.dflt = defaultButtons(bb.defaultListener)
	}
	return bb.dflt
}

func (bb *buttons) defaultListener(label string) {
	switch label {
	case "args":
		bb.viewUpd(bb.argsButtons())
	case "more":
		bb.viewUpd(bb.moreButtons())
	}
}

func (bb *buttons) moreButtons() *buttoner {
	if bb.more == nil {
		bb.more = moreButtons(bb.moreListener)
	}
	return bb.more
}

func (bb *buttons) moreListener(label string) {
	switch label {
	case "back":
		bb.viewUpd(bb.defaultButtons())
		bb.viewUpd(bb.lastReport)
	case "help":
		viewHelp(bb.viewUpd)
	case "about":
		viewAbout(bb.viewUpd)
	}
}

func (bb *buttons) argsButtons() *buttoner {
	if bb.args == nil {
		bb.args = argsButtons(bb.argsListener)
	}
	return bb.args
}

func (bb *buttons) argsListener(label string) {
	switch label {
	case "back":
		bb.viewUpd(bb.defaultButtons())
		bb.viewUpd(bb.lastReport)
	}
}

const (
	bttPkgs  = "pkgs"
	bttSuits = "suites=off"
	bttArgs  = "args"
	bttMore  = "more"
)

func defaultButtons(l func(string)) *buttoner {
	return &buttoner{
		replace:  true,
		listener: l,
		newBB: []view.ButtonDef{
			{Label: bttPkgs, Rune: 'p'},
			{Label: bttSuits, Rune: 's'},
			{Label: bttArgs, Rune: 'a'},
			{Label: bttMore, Rune: 'm'},
		},
	}
}

func moreButtons(l func(string)) *buttoner {
	return &buttoner{
		replace:  true,
		listener: l,
		newBB: []view.ButtonDef{
			{Label: "help", Rune: 'h'},
			{Label: "about", Rune: 'a'},
			{Label: "quit", Rune: 'q'},
			{Label: "back", Rune: 'b'},
		},
	}
}

func argsButtons(l func(string)) *buttoner {
	return &buttoner{
		replace:  true,
		listener: l,
		newBB: []view.ButtonDef{
			{Label: "race=off", Rune: 'r'},
			{Label: "vet=off", Rune: 'v'},
			{Label: "stats=off", Rune: 's'},
			{Label: "back", Rune: 'b'},
		},
	}
}

type buttoner struct {
	replace  bool
	newBB    []view.ButtonDef
	updBB    map[string]view.ButtonDef
	listener view.ButtonLst
}

func (b buttoner) Replace() bool { return b.replace }

func (b *buttoner) ForUpdate(cb func(string, view.ButtonDef) error) {
	for label, def := range b.updBB {
		if err := cb(label, def); err != nil {
			panic(err)
		}
	}
}

func (b *buttoner) ForNew(cb func(view.ButtonDef) error) {
	for _, def := range b.newBB {
		if err := cb(def); err != nil {
			panic(err)
		}
	}
}

func (b *buttoner) Listener() view.ButtonLst {
	return b.listener
}
