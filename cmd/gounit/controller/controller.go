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
	"github.com/slukits/gounit/cmd/gounit/model"
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
	ee(nil).Listen()
}
