// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
fx_controller_test provides test-fixtures for end to end tests.  The
main difficulty is that model and view do their work in their own go
routines, i.e. it is unclear when testing packages are reported and
their reporting has been processed as well as when a view update has
made it to the screen.  The two methods beforeWatch and beforeView of
the Testing-type provide a solution since they are guaranteed to not
return before reporting packages as consequence of a provided function
call has been processed; respectively to not return before a view update
has made it to the screen as consequence of a provided function call.
*/

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

	// gounit _controller instance created during the initialization
	// process, i.e. _controller.New.
	_controller *controller

	// _beforeWatch is closed and recreated each time reported testing
	// packages have been processed.
	_beforeWatch chan struct{}

	// _beforeView is closed and recreated each time a view is update
	// has been processed.
	_beforeView chan struct{}

	// _beforeTimeout is the time span Testing's before*-methods waits
	// for either the processing of a reported testing package or the
	// update of the view.
	_beforeTimeout time.Duration

	// _quitWatching terminates the model's go routine checking for
	// updated packages in given source directory.
	_quitWatching interface{ QuitAll() }

	golden *tfs.Dir
}

func (tt *Testing) collapseAll() {
	tt._controller.model.report(rprPackages)
}

// isOn returns true if given button-mask represents on/off-buttons
// which are switched to on.
func (tt *Testing) isOn(o onMask) bool {
	return tt._controller.bb.isOn&o == o
}

// switchButtonLabel returns the currently set switch button label as it
// appears in the view, derived from given short label like "vet",
// "race" or "stats".
func (tt *Testing) switchButtonLabel(shortLabel string) (string, string) {
	bb := switchButtons(tt._controller.bb.isOn, nil)
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

// ClickButtons is a short cut for several subsequent ClickButton-calls.
func (tt *Testing) ClickButtons(ll ...string) {
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
	c <-chan *model.PackagesDiff, m *modelState, _ chan bool,
) {
	watchRelay := make(chan *model.PackagesDiff)
	afterWatch := make(chan bool)
	m.replaceViewUpdater(
		func(vu func(...interface{})) func(i ...interface{}) {
			return func(i ...interface{}) {
				// it seems that the created golden tmp-directory by
				// testing.Cleanup is removed before the closing of the
				// PackagesDiff-channel is executed hence this final
				// source-directory change is reported which then leads
				// to an error in the view update since the test run has
				// already ended when that update reaches the view.
				// Since at this point in time so far the PackagesDiff
				// channel has been already closed we leverage this to
				// suppress that final view update.
				select {
				case pd := <-c:
					if pd == nil {
						return
					}
				default:
				}
				vu(i...)
				tt.Lock()
				close(tt._beforeView)
				tt._beforeView = make(chan struct{})
				tt.Unlock()
			}
		}(m.viewUpdater))
	go watch(watchRelay, m, afterWatch)
	for pd := range c {
		if pd == nil {
			close(watchRelay)
			return
		}
		watchRelay <- pd
		<-afterWatch
		tt.Lock()
		close(tt._beforeWatch)
		tt._beforeWatch = make(chan struct{})
		tt.Unlock()
	}
}

type uiCmp int

const (
	awMessageBar uiCmp = iota
	awReporting
	awStatusBar
	awButtons
)

// beforeView executes given function and waits for a view update.
// beforeView fatales wrapped test if a set timeout occurs before a view
// update happened.
func (tt *Testing) beforeView(f func()) {
	tt.T.GoT().Helper()
	tt.Lock()
	cn := tt._beforeView
	tt.Unlock()
	f()
	select {
	case <-cn:
	case <-tt.T.Timeout(tt._beforeTimeout):
		tt.T.Fatal("controller: testing: before view: " +
			"timed out without a view-update")
	}
}

// beforeWatch calls given function and waits for an update of watched
// directory, i.e. given function must trigger an source-file update in
// watched directory.
func (tt *Testing) beforeWatch(f func()) {
	tt.T.GoT().Helper()
	tt.Lock()
	cn := tt._beforeWatch
	tt.Unlock()
	f()
	select {
	case <-cn:
	case <-tt.T.Timeout(tt._beforeTimeout):
		tt.T.Fatal("controller: testing: before watch: " +
			"timed out without a watch-update")
	}
}

func (tt *Testing) before(f func()) {
	tt.T.GoT().Helper()
	tt.Lock()
	bw := tt._beforeWatch
	bv := tt._beforeView
	tt.Unlock()
	f()
	select {
	case <-bw:
	case <-tt.T.Timeout(tt._beforeTimeout):
		tt.T.Fatal("controller: testing: before: " +
			"timed out without a watch-update")
	}
	select {
	case <-bv:
	case <-tt.T.Timeout(tt._beforeTimeout):
		tt.T.Fatal("controller: testing: before: " +
			"timed out without a view-update")
	}
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
func fx(t *gounit.T) *Testing {
	return fxSource(t, "empty")
}

func fxDBG(t *gounit.T) *Testing {
	return fxSourceDBG(t, "empty")
}

// initGolden guarantees that the golden module is only once
// initialized.
var initGolden = func(done bool) func(t tfs.Tester) {
	mutex := &sync.Mutex{}
	return func(t tfs.Tester) {
		t.GoT().Helper()
		mutex.Lock()
		defer mutex.Unlock()
		if done {
			return
		}
		done = true
		dt, _ := t.FS().Data()
		golden := dt.Child(goldenDir)
		os.Remove(fp.Join(golden.Path(), "go.mod"))
		golden.MkMod("example.com/gounit/controller/golden")
		golden.MkTidy()
	}
}(false)

func fxSource(t *gounit.T, relDir string) *Testing {
	golden := fxSetupSource(t, relDir)
	return fxInit(
		t,
		InitFactories{
			Watcher: &model.Sources{
				Dir:      fp.Join(golden.Path(), relDir),
				Interval: 1 * time.Millisecond,
			},
		},
		golden,
	)
}

func fxSourceTouched(
	t *gounit.T, relDir string, touch string,
) *Testing {
	golden := fxSetupSource(t, relDir)
	time.Sleep(1 * time.Millisecond)
	golden.Touch(touch)
	return fxInit(
		t,
		InitFactories{
			Watcher: &model.Sources{
				Dir:      fp.Join(golden.Path(), relDir),
				Interval: 1 * time.Millisecond,
			},
		},
		golden,
	)
}

func fxSourceDBG(t *gounit.T, relDir string) *Testing {
	golden := fxSetupSource(t, relDir)
	return fxInit(
		t,
		InitFactories{
			dbgTimeouts: true,
			Watcher: &model.Sources{
				Dir:      fp.Join(golden.Path(), relDir),
				Interval: 1 * time.Millisecond,
			},
		},
		golden,
	)
}

// fxSourceAbsDBG sets up a testing fixture for debugging using given
// absolute directory as watched source directory.  This allows to test
// and debug against go modules from the wild.
func fxSourceAbsDBG(t *gounit.T, absDir string) *Testing {
	return fxInit(
		t,
		InitFactories{
			dbgTimeouts: true,
			Watcher: &model.Sources{
				Dir:      absDir,
				Interval: 1 * time.Millisecond,
			},
		},
		nil,
	)
}

func fxSourceTouchedDBG(
	t *gounit.T, relDir, touch string,
) *Testing {
	golden := fxSetupSource(t, relDir)
	time.Sleep(1 * time.Millisecond)
	golden.Touch(touch)
	return fxInit(
		t,
		InitFactories{
			dbgTimeouts: true,
			Watcher: &model.Sources{
				Dir:      fp.Join(golden.Path(), relDir),
				Interval: 1 * time.Millisecond,
			},
		},
		golden,
	)
}

func fxSetupSource(t *gounit.T, relDir string) (golden *tfs.Dir) {
	t.GoT().Helper()
	tmp := t.FS().Tmp()
	dt, _ := t.FS().Data()
	golden = dt.Child(goldenDir)
	if _, err := os.Stat(golden.Path()); err != nil {
		t.Fatalf("fx: watch: testdata: golden-module: %v", err)
	}
	golden.Copy(tmp)
	golden = tmp.Child(goldenDir)
	if _, err := os.Stat(golden.Path()); err != nil {
		t.Fatalf("fx: watch: tmp-copy: golden-module: %v", err)
	}
	_, err := os.Stat(fp.Join(golden.Path(), relDir))
	if err != nil {
		t.Fatalf("fx: watch: %s: %v", relDir, err)
	}
	return golden
}

// fxInit creates controller test fixture providing the controllers
// Events instance and its testing instance.  fxInit guarantees to no
// return before the initial report of the first model-report about the
// watched source directory.  I.e. tests on an non-testing source
// directory may not used this fixture factory.
func fxInit(
	t *gounit.T, i InitFactories, golden *tfs.Dir,
) *Testing {

	var ct Testing
	ct.Mutex = &sync.Mutex{}
	ct._beforeTimeout = 30 * time.Second
	if i.dbgTimeouts {
		ct._beforeTimeout = 20 * time.Minute
	}
	ct._beforeWatch = make(chan struct{})
	ct._beforeView = make(chan struct{})
	ct.golden = golden

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
	i.Events = func(c lines.Componenter) *lines.Lines {
		timeout := 10 * time.Second
		if i.dbgTimeouts {
			timeout = 20 * time.Minute
		}
		ct.Testing = view.FixtureFor(t, timeout, c)
		ct.Lines.OnQuit(func() {
			if ct._quitWatching != nil {
				ct._quitWatching.QuitAll()
			}
		})
		return ct.Lines
	}

	ensureInit := ct._beforeWatch
	New(&i)

	ct._controller = i.controller
	if qw, ok := i.Watcher.(interface{ QuitAll() }); ok {
		t.GoT().Cleanup(func() { qw.QuitAll() })
	}
	if !strings.HasSuffix(i.Watcher.SourcesDir(), "empty") {
		<-ensureInit
	}
	return &ct
}

// watcherMock mocks up the controller.New Watcher argument.
type watcherMock struct {
	c     chan *model.PackagesDiff
	watch func() (<-chan *model.PackagesDiff, uint64, error)
}

const (
	mckModule    = "mock-module"
	mckModuleDir = "mock/module/dir"
	mckSourceDir = "mock/source/empty"
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
	ee *lines.Lines
	tt *lines.Fixture
}

// fixtureSetter provides implements a method to store a fixture which
// needs cleaning up usually a gounit.Fixtures instance.
type fixtureSetter interface {
	Set(*gounit.T, interface{})
}
