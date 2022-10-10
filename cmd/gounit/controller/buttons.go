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
	cls        *buttoner
	isOn       onMask
	quitter    func()
}

func newButtons(upd func(...interface{})) *buttons {
	return &buttons{viewUpd: upd}
}

func (bb *buttons) defaults() *buttoner {
	if bb.dflt == nil {
		bb.dflt = defaultButtons(bb.defaultListener)
	}
	return bb.dflt
}

func (bb *buttons) defaultListener(label string) {
	switch label {
	case "close":
		bb.viewUpd(bb.defaults())
		bb.modelState.resume()
	case "switches":
		bb.viewUpd(bb.switchButtons())
	case "help":
		bb.modelState.suspend()
		bb.viewUpd(viewHelp(), bb.close())
	case "quit":
		bb.quitter()
	case "about":
		bb.modelState.suspend()
		bb.viewUpd(viewAbout(), bb.close())
	}
}

func (bb *buttons) close() *buttoner {
	if bb.cls == nil {
		bb.cls = &buttoner{
			replace:  true,
			listener: bb.defaultListener,
			newBB:    []view.ButtonDef{{Label: "close", Rune: 'c'}},
		}
	}
	return bb.cls
}

func (bb *buttons) switchButtons() *buttoner {
	return switchButtons(bb.isOn, bb.switchesListener)
}

func (bb *buttons) switchesListener(label string) {
	switch label {
	case "back":
		bb.viewUpd(bb.defaults())
	case bttVetOff:
		bb.isOn |= vetOn
		bb.viewUpd(switchButtons(bb.isOn, bb.switchesListener))
		bb.modelState.setOnFlag(vetOn)
	case bttVetOn:
		bb.isOn &^= vetOn
		bb.viewUpd(switchButtons(bb.isOn, bb.switchesListener))
		bb.modelState.removeOneFlag(vetOn)
	case bttRaceOff:
		bb.isOn |= raceOn
		bb.viewUpd(switchButtons(bb.isOn, bb.switchesListener))
		// TODO: removing this line we get a wired error report go figure
		bb.modelState.setOnFlag(raceOn)
	case bttRaceOn:
		bb.isOn &^= raceOn
		bb.viewUpd(switchButtons(bb.isOn, bb.switchesListener))
		bb.modelState.removeOneFlag(raceOn)
	case bttStatsOff:
		bb.isOn |= statsOn
		bb.viewUpd(switchButtons(bb.isOn, bb.switchesListener))
		bb.modelState.setOnFlag(statsOn)
	case bttStatsOn:
		bb.isOn &^= statsOn
		bb.viewUpd(switchButtons(bb.isOn, bb.switchesListener))
		bb.modelState.removeOneFlag(statsOn)
	}
}

func defaultButtons(l func(string)) *buttoner {
	return &buttoner{
		replace:  true,
		listener: l,
		newBB: []view.ButtonDef{
			{Label: "switches", Rune: 's'},
			{Label: "help", Rune: 'h'},
			{Label: "about", Rune: 'a'},
			{Label: "quit", Rune: 'q'},
		},
	}
}

const (
	bttRaceOff    = "race=off"
	bttRaceOn     = "race=on"
	bttVetOff     = "vet=off"
	bttVetOn      = "vet=on"
	bttStatsOff   = "stats=off"
	bttStatsOn    = "stats=on"
	bttCurrentOff = "current=off"
	bttCurrentOn  = "current=on"
)

func switchButtons(on onMask, l func(string)) *buttoner {
	bb := &buttoner{
		replace:  true,
		listener: l,
		newBB: []view.ButtonDef{
			{Label: bttVetOff, Rune: 'v'},
			{Label: bttRaceOff, Rune: 'r'},
			{Label: bttStatsOff, Rune: 's'},
			{Label: bttCurrentOff, Rune: 'c'},
			{Label: "back", Rune: 'b'},
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
