// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"os"
	fp "path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/slukits/gounit"
	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/gounit/pkg/tfs"
	"github.com/slukits/lines"
)

// A Testing instance provides conveniences for controller-testing.
type Testing struct {

	// testing instance of gounit's view.
	*view.Testing

	// mutex protecting the access and update of afterWatch.
	*sync.Mutex

	// testing lines events instance.
	_ee *lines.Events

	// gounit _controller instance created during the initialization
	// porcess, i.e. _controller.New.
	_controller *controller

	// this channel is recreated each time a source directory update is
	// reported and closed after an subsequent update of the view.
	_afterWatch chan struct{}

	// _watchTimeout is the time span Testing's AfterWatch-method waits
	// for an update of the view.  Practically that means the time span
	// between a reported packages of a source watcher and the update in
	// the view.  Since the execution of the tests happen in between
	// this time-span can't be to short (> 0.2sec).
	_watchTimeout time.Duration

	// _quitWatching terminates the model's go routine checking for
	// updated packages in given source directory.
	_quitWatching interface{ QuitAll() }
}

// cleanUp stop all go-routines initiated by a controller test.
func (tt *Testing) cleanUp() {
	if tt._ee != nil && tt._ee.IsListening() {
		tt._ee.QuitListening()
	}
	if tt._quitWatching != nil {
		tt._quitWatching.QuitAll()
	}
}

// isOn returns true if given button-mask represents on/off-buttons
// which are switched to on.
func (tt *Testing) isOn(o onMask) bool {
	return tt._controller.bb.isOn&o == o
}

// dfltButtonLabel returns the currently set argument button label as it
// appears in the view, derived from given short label like "vet",
// "race" or "stats".
func (tt *Testing) dfltButtonLabel(shortLabel string) (string, string) {
	bb := defaultButtons(tt._controller.bb.isOn, nil)
	lbl, vw := "", ""
	bb.ForNew(func(bd view.ButtonDef) error {
		if lbl != "" {
			return nil
		}
		if !strings.HasPrefix(bd.Label, shortLabel) {
			return nil
		}
		lbl = bd.Label
		vw = strings.Replace(bd.Label, string(bd.Rune),
			fmt.Sprintf("[%c]", bd.Rune), 1)
		return nil
	})
	return lbl, vw
}

// splitTrimmed splits given string at line-breaks and trims resulting
// lines of leading or trailing whitespace.
func (tt *Testing) splitTrimmed(s string) []string {
	ss := strings.Split(s, "\n")
	for i, l := range ss {
		ss[i] = strings.TrimSpace(l)
	}
	return ss
}

// clickButtons is a short cut for several subsequent ClickButton-calls.
func (tt *Testing) clickButtons(ll ...string) {
	tt.T.GoT().Helper()
	for _, l := range ll {
		tt.ClickButton(l)
	}
}

// _fxWatchMock is started as go-routine during the controller's
// initialization process in controller.New and wraps the default watch
// function.  The later reports model changes (i.e. watched packages
// changes) to the view.  _fxWatchMock intercepts that report to close
// and replace the current tt._afterWatch channel after the view update
// was processed.  This enables a test to wait on afterWatch to ensure
// the view and screen update process has completed.
func (tt *Testing) _fxWatchMock(
	c <-chan *model.PackagesDiff, m *modelState,
) {
	watchRelay := make(chan *model.PackagesDiff)
	m.replaceViewUpdater(
		func(vu func(...interface{})) func(i ...interface{}) {
			return func(i ...interface{}) {
				vu(i...)
				tt.Lock()
				close(tt._afterWatch)
				tt._afterWatch = make(chan struct{})
				tt.Unlock()
			}
		}(m.viewUpdater))
	go watch(watchRelay, m)
	for pd := range c {
		if pd == nil {
			close(watchRelay)
			return
		}
		watchRelay <- pd
	}
}

type uiCmp int

const (
	awMessageBar uiCmp = iota
	awReporting
	awStatusBar
	awButtons
)

// afterWatch returns the requested screen portion after the watcher of
// a source directory has reported a change to the view which in turn
// has updated the screen.
func (tt *Testing) afterWatch(c uiCmp) lines.TestScreen {
	tt.T.GoT().Helper()
	tt.Lock()
	cn := tt._afterWatch
	tt.Unlock()
	select {
	case <-cn:
		switch c {
		case awMessageBar:
			return tt.MessageBar()
		case awReporting:
			return tt.Reporting()
		case awStatusBar:
			return tt.StatusBar()
		case awButtons:
			return tt.ButtonBar()
		}
	case <-tt.T.Timeout(tt._watchTimeout):
		tt.T.Fatal("controller: testing: after watch: " +
			"timed out without a watch-update")
	}
	return nil
}

const (
	goldenDir    = "goldenmod"
	goldenModule = "example.com/gounit/controller/golden"
	emptyPkg     = "empty"
)

// fx creates a new controller fixture and returns the lines.Events and
// controller.Testing instances instantiated by the controller.  If a
// fixtureSetter is given a cleanup method is stored to given fixture
// setter. NOTE fx doesn't return before controller.New is listening.
func fx(t *gounit.T, fs fixtureSetter) (*lines.Events, *Testing) {
	return fxSource(t, fs, "empty")
}

func fxDBG(t *gounit.T, fs fixtureSetter) (*lines.Events, *Testing) {
	return fxSourceDBG(t, fs, "empty")
}

func initGolden(t tfs.Tester) {
	dt, _ := t.FS().Data()
	golden := dt.Child(goldenDir)
	if _, err := os.Stat(fp.Join(golden.Path(), "go.mod")); err != nil {
		golden.MkMod("example.com/gounit/controller/golden")
	}
	golden.MkTidy()
}

func fxSource(t *gounit.T, fs fixtureSetter, relDir string) (
	*lines.Events, *Testing,
) {
	golden := fxSetupSource(t, relDir)
	return fxInit(
		t,
		fs,
		InitFactories{
			Watcher: &model.Sources{
				Dir:      fp.Join(golden.Path(), relDir),
				Interval: 1 * time.Millisecond,
			},
		},
	)
}

func fxSourceDBG(t *gounit.T, fs fixtureSetter, relDir string) (
	*lines.Events, *Testing,
) {
	golden := fxSetupSource(t, relDir)
	return fxInit(
		t,
		fs,
		InitFactories{
			dbgTimeouts: true,
			Watcher: &model.Sources{
				Dir:      fp.Join(golden.Path(), relDir),
				Interval: 1 * time.Millisecond,
			},
		},
	)
}

func fxSetupSource(t *gounit.T, relDir string) (golden *tfs.Dir) {
	tmp := t.FS().Tmp()
	dt, _ := t.FS().Data()
	golden = dt.Child(goldenDir)
	golden.Copy(tmp)
	golden = tmp.Child(goldenDir)
	_, err := os.Stat(fp.Join(golden.Path(), relDir))
	if err != nil {
		t.Fatalf("fx: watch: %s: %v", relDir, err)
	}
	return golden
}

// fxInit is like fx but also takes an instance of init factories
func fxInit(t *gounit.T, fs fixtureSetter, i InitFactories) (
	*lines.Events, *Testing,
) {

	var (
		ct Testing
		ee *lines.Events
	)
	ct.Mutex = &sync.Mutex{}
	ct._watchTimeout = 2 * time.Second
	if i.dbgTimeouts {
		ct._watchTimeout = 20 * time.Minute
	}
	ct._afterWatch = make(chan struct{})

	if i.Fatal == nil {
		i.Fatal = func(i ...interface{}) {
			t.Fatalf("unexpected error: %s", fmt.Sprint(i...))
		}
	}
	if i.Watcher == nil {
		i.Watcher = &watcherMock{}
	}
	if i.watch == nil {
		i.watch = ct._fxWatchMock
	}
	i.Events = func(c lines.Componenter) *lines.Events {
		events, tt := lines.Test(t.GoT(), c)
		if i.dbgTimeouts {
			tt.Timeout = 20 * time.Minute
		}
		ct.Testing = view.NewTesting(t, tt, c)
		ee = events
		ct._ee = events
		return ee
	}

	New(&i)

	ct._controller = i.controller
	if qw, ok := i.Watcher.(interface{ QuitAll() }); ok {
		ct._quitWatching = qw
	}
	if fs != nil {
		fs.Set(t, ct.cleanUp)
	}
	return ee, &ct
}

// watcherMock mocks up the controller.New Watcher argument.
type watcherMock struct {
	c     chan *model.PackagesDiff
	watch func() (<-chan *model.PackagesDiff, uint64, error)
}

const (
	mckModule    = "mock-module"
	mckModuleDir = "mock/module/dir"
	mckSourceDir = "mock/source/dir"
)

func (m *watcherMock) ModuleName() string { return mckModule }
func (m *watcherMock) ModuleDir() string  { return mckModuleDir }
func (m *watcherMock) SourcesDir() string { return mckSourceDir }
func (m *watcherMock) Watch() (
	<-chan *model.PackagesDiff, uint64, error,
) {
	if m.watch != nil {
		return m.watch()
	}
	m.c = make(chan *model.PackagesDiff)
	return m.c, 1, nil
}

type linesTest struct {
	ee *lines.Events
	tt *lines.Testing
}

// fixtureSetter provides implements a method to store a fixture which
// needs cleaning up usually a gounit.Fixtures instance.
type fixtureSetter interface {
	Set(*gounit.T, interface{})
}
