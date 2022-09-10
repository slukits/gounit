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
	"strings"
	"sync"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

// InitFactories allows to overwrite the default constructors and
// functionality needed to initialize the controller.
type InitFactories struct {

	// Fatal to report fatal errors; defaults to log.Fatal
	Fatal func(...interface{})

	// Watcher wraps a controller's model; defaults to &model.Sources{}
	Watcher Watcher

	// View returns a controller's view; defaults to view.New
	View func(view.Initer) lines.Componenter

	// Events returns a controller's events loop; defaults to lines.New
	Events func(lines.Componenter) *lines.Events
}

// New starts the application and blocks until a quit event occurs.
func New(i InitFactories) {
	ensureInitArgs(&i)
	diff, _, err := i.Watcher.Watch()
	if err != nil {
		i.Fatal(WatcherErr, i.Watcher.SourcesDir(), err)
		return
	}
	_ = diff
	vwInit := &vwIniter{vw: &vwUpd{mutex: &sync.Mutex{}}}
	vwInit.w, vwInit.ftl = i.Watcher, i.Fatal
	ee := i.Events(i.View(vwInit))
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
}

// vwUpd collects the functions to update aspects of a view which are
// used by the viewUpdater to modify the view.
type vwUpd struct {
	// mutex avoids that the view is updated concurrently.
	mutex *sync.Mutex

	// msg updates the view's message bar
	msg func(string)

	// stt updates the view's status bar
	stt func(view.StatusUpdate)

	// rprUpd updates lines of a view's reporting component.
	rprUpd view.ReportingUpd

	// bttUpd updates the buttons of a view's button bar.
	bttUpd func(view.Buttoner)
}

// Update updates the view and should be the only way the view is
// updated to avoid data races.
func (vw *vwUpd) Update(data interface{}) {
	vw.mutex.Lock()
	defer vw.mutex.Unlock()

	switch updData := data.(type) {
	case view.Buttoner:
		vw.bttUpd(updData)
	case view.Liner:
		vw.rprUpd(updData)
	}
}

// viewUpdater runs concurrently and is kicked of by the controller
// constructor.  All updates of the view should be send to its update
// channel upd.
func viewUpdater(vw *vwUpd, upd <-chan interface{}) {
	for {
		updData := <-upd
		if updData == nil {
			return
		}
	}
}

const initReport = "waiting for testing packages being reported..."

// vwIniter instance provides the initial data to a new view and
// collects the provided view modifiers.
type vwIniter struct {
	w   Watcher
	ftl func(...interface{})
	vw  *vwUpd
}

func (i *vwIniter) Fatal() func(...interface{}) { return i.ftl }

func (i *vwIniter) Message(msg func(string)) string {
	i.vw.msg = msg
	return fmt.Sprintf("%s: %s",
		i.w.ModuleName(), strings.TrimPrefix(
			i.w.SourcesDir(), i.w.ModuleDir()))
}

func (i *vwIniter) Reporting(
	upd view.ReportingUpd,
) (string, view.ReportingLst) {

	i.vw.rprUpd = upd
	return initReport, func(idx int) {}
}

func (i *vwIniter) Status(upd func(view.StatusUpdate)) {
	i.vw.stt = upd
}

func (i *vwIniter) Buttons(upd func(view.Buttoner)) view.Buttoner {
	i.vw.bttUpd = upd
	return newButtons(
		i.vw.Update, &liner{clearing: true, ll: []string{initReport}},
	).defaultButtons()
}

const (
	WatcherErr = "gounit: watcher: %s: %v"
)

// A Watcher implementation provides the needed information about a
// watched source directory to the controller.
type Watcher interface {

	// ModuleName is the module name of the descendant watched source
	// directory.
	ModuleName() string

	// ModuleDir returns the absolute path of the module directory of
	// the descendant watched source directory.
	ModuleDir() string

	// SourcesDir returns the absolute path of the watched source
	// directory.
	SourcesDir() string

	// Watch is a function whose returned channel watches a go modules
	// packages sources whose tests runs are reported to a terminal ui.
	Watch() (<-chan *model.PackagesDiff, uint64, error)
}

type liner struct {
	clearing bool
	ll       []string
	mask     map[uint]view.LineMask
}

func (l *liner) Clearing() bool { return l.clearing }

func (l *liner) For(_ lines.Componenter, line func(uint, string)) {
	for idx, content := range l.ll {
		line(uint(idx), content)
	}
}

func (l *liner) Mask(idx uint) view.LineMask { return l.mask[idx] }
