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
	ee *lines.Events
	// gounit view's buttons.
	bb *buttons
	// this channel is recreated each time a source directory update is
	// reported and closed after an subsequent update of the view.
	afterWatch chan struct{}
	// watchTimeout is the time span Testing's AfterWatch-method waits
	// for an update of the view.  Practically that means the time span
	// between a reported packages of a source watcher and the update in
	// the view.  Since the execution of the tests happen in between
	// this time-span can't be to short (> 0.2sec).
	watchTimeout time.Duration

	quitWatching interface{ QuitAll() }
}

// cleanUp stop all go-routines initiated by a controller test.
func (tt *Testing) cleanUp() {
	if tt.ee != nil && tt.ee.IsListening() {
		tt.ee.QuitListening()
	}
	if tt.quitWatching != nil {
		tt.quitWatching.QuitAll()
	}
}

// isOn returns true if given button-mask represents on/off-buttons
// which are switched to on.
func (tt *Testing) isOn(o onMask) bool {
	return tt.bb.isOn&o == o
}

// dfltButtonLabel returns the currently set argument button label as it
// appears in the view, derived from given short label like "vet",
// "race" or "stats".
func (tt *Testing) dfltButtonLabel(shortLabel string) (string, string) {
	bb := defaultButtons(tt.bb.isOn, nil)
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

func (tt *Testing) fxWatchMock(
	c <-chan *model.PackagesDiff, f func(...interface{}),
) {
	watchRelay := make(chan *model.PackagesDiff)
	go watch(watchRelay, func(i ...interface{}) {
		f(i...)
		tt.Lock()
		close(tt.afterWatch)
		tt.afterWatch = make(chan struct{})
		tt.Unlock()
	})
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

// AfterWatch returns the requested screen portion after the watcher of
// a source directory has updated the screen.
func (tt *Testing) AfterWatch(c uiCmp) lines.TestScreen {
	tt.T.GoT().Helper()
	tt.Lock()
	cn := tt.afterWatch
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
	case <-tt.T.Timeout(tt.watchTimeout):
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

// fxInit is like fx but also takes an instance of init fac
func fxInit(t *gounit.T, fs fixtureSetter, i InitFactories) (
	*lines.Events, *Testing,
) {

	var (
		ct Testing
		ee *lines.Events
	)
	ct.Mutex = &sync.Mutex{}
	ct.watchTimeout = 2 * time.Second
	if i.dbgTimeouts {
		ct.watchTimeout = 20 * time.Minute
	}
	ct.afterWatch = make(chan struct{})

	if i.Fatal == nil {
		i.Fatal = func(i ...interface{}) {
			t.Fatalf("unexpected error: %s", fmt.Sprint(i...))
		}
	}
	if i.Watcher == nil {
		i.Watcher = &watcherMock{}
	}
	if i.watch == nil {
		i.watch = ct.fxWatchMock
	}
	i.View = func(i view.Initer) lines.Componenter {
		vw := view.New(i)
		ct.bb = i.(*vwIniter).bb
		return vw
	}
	i.Events = func(c lines.Componenter) *lines.Events {
		events, tt := lines.Test(t.GoT(), c)
		if i.dbgTimeouts {
			tt.Timeout = 20 * time.Minute
		}
		ct.Testing = view.NewTesting(t, tt, c)
		ee = events
		ct.ee = events
		return ee
	}

	New(i)

	if qw, ok := i.Watcher.(interface{ QuitAll() }); ok {
		ct.quitWatching = qw
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
