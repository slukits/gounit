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
sources directory.  Further more a suite variant of the default or
package view can be requested which shows all the "current" package's
testing suits.  Finally the user can choose to switch on/off: race, vet
and stats.  The later tells how many source files are in the current
package/watched sources, how many of them are testing files how many
lines of code, how many of them are for testing and how many lines of
documentation were found.
*/
package controller

import (
	"fmt"
	"log"
	"sync"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

// InitFactories allows to overwrite the default constructors and
// functionality needed to initialize the controller and to obtain the
// created controller for testing purposes.
type InitFactories struct {

	// Fatal to report fatal errors; defaults to log.Fatal
	Fatal func(...interface{})

	// Watcher wraps a controller's model; defaults to &model.Sources{}
	Watcher Watcher

	// watch waits concurrently for a watcher to report watched testing
	// packages and updates accordingly the view; defaults to a
	// controller internal function and is there for testing
	// interceptions.
	watch func(<-chan *model.PackagesDiff, *modelState)

	// dbgTimeouts set to true increases lines.Events.sync-timeout to
	// 20min as well as controller.Testing.watchTimeout.
	dbgTimeouts bool

	// controller instance setup during the initialization process.
	controller *controller

	// View returns a controller's view; defaults to view.New
	View func(view.Initer) lines.Componenter

	// Events returns a controller's events loop; defaults to lines.New
	Events func(lines.Componenter) *lines.Events
}

type controller struct {
	view    *viewUpdater
	model   *modelState
	watcher Watcher
	bb      *buttons
	ftl     func(...interface{})
}

// New starts the application and blocks until a quit event occurs.
// Providing the zero-InitFactories uses the documented defaults (see
// [controller.InitFactories]).
func New(i *InitFactories) {
	if i == nil {
		i = &InitFactories{}
	}
	ensureInitArgs(i)
	diff, _, err := i.Watcher.Watch()
	if err != nil {
		i.Fatal(fmt.Sprintf(
			WatcherErr, i.Watcher.SourcesDir(), err))
		return
	}
	ee := i.Events(i.View(&viewIniter{controller: i.controller}))
	i.controller.bb.quitter = ee.QuitListening
	i.controller.bb.reporter = i.controller.model.report
	go i.watch(diff, i.controller.model)
	ee.Listen()
}

func ensureInitArgs(i *InitFactories) {
	if i.Fatal == nil {
		i.Fatal = log.Fatal
	}
	if i.Watcher == nil {
		i.Watcher = &model.Sources{}
	}
	if i.View == nil {
		i.View = func(i view.Initer) lines.Componenter {
			return view.New(i)
		}
	}
	if i.Events == nil {
		i.Events = lines.New
	}
	if i.watch == nil {
		i.watch = watch
	}

	i.controller = &controller{
		view: &viewUpdater{Mutex: &sync.Mutex{}},
		model: &modelState{
			Mutex: &sync.Mutex{},
			lastReport: []interface{}{&report{
				ll: []string{initReport}, flags: view.RpClearing}},
			pp: pkgs{},
		},
		watcher: i.Watcher,
		ftl:     i.Fatal,
	}
	i.controller.model.viewUpdater = i.controller.view.Update
	i.controller.bb = newButtons(i.controller.view.Update)
}
