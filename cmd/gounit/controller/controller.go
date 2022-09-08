// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package controller starts gounit's event loop and connects the model
with the view by feeding requested information extracted from watched
sources to the view.  Initially the controller reports information about
the most recently modified package p of the watched sources by feeding
the test results of the last suite of the most recently modified
*_test.go file of p to the view as well as a summary about the tests of
p.  This behavior is superseded by one or more failing tests.  Is there
one failing test, then its package and its failing suite or the package
tests including the failing test are reported.  Is there more than one
failing test only the failing tests are shown with their suite and
package information.  After a watch source file changes its package
becomes the currently reported package and its latest modified
suite/test is shown.

A user can request the package overview showing all packages of a watch
sources directory.  Further more a suite view can be requested which
shows all the "current" package's testing suits.  Finally the user can
choose to switch on/off: race, vet and stats.  The later tells how many
source files are in the current package/watched sources, how many of
them are testing files how many lines of code, how many of them are for
testing and how many lines of documentation were found.
*/
package controller

import (
	"fmt"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

const (
	WatcherErr = "gounit: watcher: %s: %v"
)

// Events is a function to initialize the terminal ui with components
// returning an Events-instance to listen for events.
type Events func(lines.Componenter) *lines.Events

// Watcher is a function whose returned channel watches a go modules
// packages sources whose tests runs are reported to a terminal ui.
type Watcher interface {
	ModuleName() string
	ModuleDir() string
	SourcesDir() string
	Watch() (<-chan *model.PackagesDiff, uint64, error)
}

// vwUpd collects the function to update aspects of a view.
type vwUpd struct {
	*lines.Events
	// msg updates the view's message bar
	msg func(string)
	// stt updates the view's status bar
	stt func(view.StatusUpdate)
	// rprUpd updates lines of a view's reporting component.
	rprUpd view.ReportingUpd
	// bttUpd updates the buttons of a view's button bar.
	bttUpd func(view.ButtonDef, view.ButtonUpdater) error
}

func (u *vwUpd) buttonListener(label string) {
	switch label {
	case "quit":
		u.QuitListening()
	}
}

type vwIniter struct {
	w   Watcher
	ftl func(...interface{})
	upd *vwUpd
}

func (i *vwIniter) Fatal() func(...interface{}) { return i.ftl }

func (i *vwIniter) Message(msg func(string)) string {
	i.upd.msg = msg
	return fmt.Sprintf("%s: %s",
		i.w.ModuleName(), i.w.SourcesDir()[len(i.w.ModuleDir()):])
}

func (i *vwIniter) Reporting(
	upd view.ReportingUpd,
) (string, view.ReportingLst) {

	i.upd.rprUpd = upd
	return "waiting for testing packages being reported...",
		func(idx int) {}
}

func (i *vwIniter) Status(upd func(view.StatusUpdate)) {
	i.upd.stt = upd
}

func (i *vwIniter) Buttons(
	upd view.ButtonUpd, cb func(view.ButtonDef) error,
) view.ButtonLst {
	dd := []view.ButtonDef{
		{Label: "pkgs", Rune: 'p'},
		{Label: "suites", Rune: 's'},
		{Label: "settings", Rune: 't'},
		{Label: "help", Rune: 'h'},
		{Label: "quit", Rune: 'q'},
	}
	for _, def := range dd {
		if err := cb(def); err != nil {
			i.ftl(err)
		}
	}
	return i.upd.buttonListener
}

// New starts the application and blocks until a quit event occurs.
// Fatale errors are reported to ftl while ll is used to initialize the
// ui and start the event loop.
func New(ftl func(...interface{}), w Watcher, ee Events) {
	diff, _, err := w.Watch()
	if err != nil {
		ftl(WatcherErr, w.SourcesDir(), err)
		return
	}
	_ = diff
	uu := &vwUpd{}
	uu.Events = ee(view.New(&vwIniter{w: w, ftl: ftl, upd: uu}))
	uu.Listen()
}
