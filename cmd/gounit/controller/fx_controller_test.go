// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"strings"

	"github.com/slukits/gounit"
	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

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

// fx creates a new controller fixture and returns the lines.Events and
// view.Testing instances instantiated by the controller.  If a
// fixtureSetter is given the Events-instance's QuitListening method is
// stored to given fixture setter. NOTE fx doesn't return before
// controller.New is listening.
func fx(t *gounit.T, fs fixtureSetter) (*lines.Events, *Testing) {
	return fxInit(t, fs, InitFactories{})
}

type Testing struct{ *view.Testing }

// waitFor executes given function and waits for the view's string
// representation to change.
func (tt *Testing) waitFor(f func()) {
	tt.T.GoT().Helper()
	str := tt.FullScreen().String()
	f()
	tt.T.Within((&gounit.TimeStepper{}).SetDuration(tt.Timeout),
		func() bool {
			return str != tt.FullScreen().String()
		})
}

func (tt *Testing) SplitTrimmed(s string) []string {
	ss := strings.Split(s, "\n")
	for i, l := range ss {
		ss[i] = strings.TrimSpace(l)
	}
	return ss
}

func (tt *Testing) Buttons(ll ...string) {
	tt.T.GoT().Helper()
	for _, l := range ll {
		tt.waitFor(func() { tt.ClickButton(l) })
	}
}

func fxInit(t *gounit.T, fs fixtureSetter, i InitFactories) (
	*lines.Events, *Testing,
) {

	var (
		ct    Testing
		ee    *lines.Events
		vwUpd chan interface{}
	)

	if i.Fatal == nil {
		i.Fatal = func(i ...interface{}) {
			t.Fatalf("unexpected error: %s", fmt.Sprint(i...))
		}
	}
	if i.Watcher == nil {
		i.Watcher = &watcherMock{}
	}
	i.View = func(i view.Initer) lines.Componenter {
		vi, ok := i.(*vwIniter)
		if !ok {
			t.Fatal("fx: init: expected vwIniter; got %T", i)
		}
		vwUpd = vi.vw.upd
		return view.New(i)
	}
	i.Events = func(c lines.Componenter) *lines.Events {
		events, tt := lines.Test(t.GoT(), c, 0)
		ct.Testing = view.NewTesting(t, tt, c)
		ee = events
		return ee
	}

	New(i)

	if fs != nil {
		fs.Set(t, func() {
			if ee != nil && ee.IsListening() {
				ee.QuitListening()
			}
			if vwUpd != nil {
				close(vwUpd)
			}
		})
	}
	return ee, &ct
}
