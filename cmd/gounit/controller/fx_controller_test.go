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

type Testing struct {
	*view.Testing
	ee *lines.Events
	bb *buttons
}

func (tt *Testing) cleanUp() {
	if tt.ee != nil && tt.ee.IsListening() {
		tt.ee.QuitListening()
	}
}

func (tt *Testing) IsOn(o onMask) bool {
	return tt.bb.isOn&o == o
}

// ArgButtonLabel returns the currently set argument button label as it
// appears in the view, derived from given short label like "vet",
// "race" or "stats".
func (tt *Testing) ArgButtonLabel(shortLabel string) (string, string) {
	bb := argsButtons(tt.bb.isOn, nil)
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

func (tt *Testing) SplitTrimmed(s string) []string {
	ss := strings.Split(s, "\n")
	for i, l := range ss {
		ss[i] = strings.TrimSpace(l)
	}
	return ss
}

func (tt *Testing) ClickButtons(ll ...string) {
	tt.T.GoT().Helper()
	for _, l := range ll {
		tt.ClickButton(l)
	}
}

// fx creates a new controller fixture and returns the lines.Events and
// controller.Testing instances instantiated by the controller.  If a
// fixtureSetter is given a cleanup method is stored to given fixture
// setter. NOTE fx doesn't return before controller.New is listening.
func fx(t *gounit.T, fs fixtureSetter) (*lines.Events, *Testing) {
	return fxInit(t, fs, InitFactories{})
}

// fxInit is like fx but also takes an instance of init fac
func fxInit(t *gounit.T, fs fixtureSetter, i InitFactories) (
	*lines.Events, *Testing,
) {

	var (
		ct Testing
		ee *lines.Events
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
		vw := view.New(i)
		ct.bb = i.(*vwIniter).bb
		return vw
	}
	i.Events = func(c lines.Componenter) *lines.Events {
		events, tt := lines.Test(t.GoT(), c, 0)
		ct.Testing = view.NewTesting(t, tt, c)
		ee = events
		ct.ee = events
		return ee
	}

	New(i)

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
