// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import "github.com/slukits/gounit/cmd/gounit/view"

type onMask uint8

const (
	vetOn onMask = 1 << iota
	raceOn
	statsOn
	errOn
)

type stater interface {
	report(reportType)
	setOnFlag(onMask)
	removeOneFlag(onMask)
	suspend()
	resume()
}

type buttons struct {
	viewUpd    func(...interface{})
	modelState stater
	dflt       *buttoner
	more       *buttoner
	isOn       onMask
	quitter    func()
}

func newButtons(upd func(...interface{})) *buttons {
	return &buttons{viewUpd: upd}
}

func (bb *buttons) defaults() *buttoner {
	if bb.dflt == nil {
		bb.dflt = defaultButtons(bb.isOn, bb.defaultListener)
	}
	return bb.dflt
}

func (bb *buttons) defaultListener(label string) {
	switch label {
	case "more":
		bb.viewUpd(bb.moreButtons())
	case bttVetOff:
		bb.isOn |= vetOn
		bb.viewUpd(defaultButtons(bb.isOn, bb.defaultListener))
		bb.modelState.setOnFlag(vetOn)
	case bttVetOn:
		bb.isOn &^= vetOn
		bb.viewUpd(defaultButtons(bb.isOn, bb.defaultListener))
		bb.modelState.removeOneFlag(vetOn)
	case bttRaceOff:
		bb.isOn |= raceOn
		bb.viewUpd(defaultButtons(bb.isOn, bb.defaultListener))
		// TODO: removing this line we get a wired error report go figure
		bb.modelState.setOnFlag(raceOn)
	case bttRaceOn:
		bb.isOn &^= raceOn
		bb.viewUpd(defaultButtons(bb.isOn, bb.defaultListener))
		bb.modelState.removeOneFlag(raceOn)
	case bttStatsOff:
		bb.isOn |= statsOn
		bb.viewUpd(defaultButtons(bb.isOn, bb.defaultListener))
		bb.modelState.setOnFlag(statsOn)
	case bttStatsOn:
		bb.isOn &^= statsOn
		bb.viewUpd(defaultButtons(bb.isOn, bb.defaultListener))
		bb.modelState.removeOneFlag(statsOn)
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
		bb.viewUpd(bb.defaults())
		bb.modelState.resume()
	case "help":
		bb.modelState.suspend()
		viewHelp(bb.viewUpd)
	case "quit":
		bb.quitter()
	case "about":
		bb.modelState.suspend()
		viewAbout(bb.viewUpd)
	}
}

const (
	bttRaceOff  = "race=off"
	bttRaceOn   = "race=on"
	bttVetOff   = "vet=off"
	bttVetOn    = "vet=on"
	bttStatsOff = "stats=off"
	bttStatsOn  = "stats=on"
	bttMore     = "more"
	bttErrOff   = "err=off"
	bttErrOn    = "err=on"
)

func defaultButtons(on onMask, l func(string)) *buttoner {
	bb := &buttoner{
		replace:  true,
		listener: l,
		newBB: []view.ButtonDef{
			{Label: bttVetOff, Rune: 'v'},
			{Label: bttRaceOff, Rune: 'r'},
			{Label: bttStatsOff, Rune: 's'},
			{Label: bttMore, Rune: 'm'},
		},
	}
	if on&vetOn > 0 {
		bb.newBB[0].Label = bttVetOn
	}
	if on&raceOn > 0 {
		bb.newBB[1].Label = bttRaceOn
	}
	if on&statsOn > 0 {
		bb.newBB[2].Label = bttStatsOn
	}
	return bb
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
