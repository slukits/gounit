// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"fmt"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/module"
	"github.com/slukits/lines"
)

// Gounit tests the behavior of Controller.New which is identical with
// the behavior of main.
type Gounit struct{ Suite }

func (s *Gounit) SetUp(t *T) { t.Parallel() }

type watcherMock struct {
	watch func() (<-chan *module.PackagesDiff, uint64, error)
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
	<-chan *module.PackagesDiff, uint64, error,
) {
	if m.watch != nil {
		return m.watch()
	}
	return make(<-chan *module.PackagesDiff), 1, nil
}

func (s *Gounit) Fails_if_watching_fails(t *T) {
	mck := &watcherMock{watch: func() (<-chan *module.PackagesDiff, uint64, error) {
		return nil, 0, errors.New("mock-err")
	}}
	fatale := false

	New(func(_ ...interface{}) { fatale = true }, mck, nil)

	t.True(fatale)
}

func mockLinesNew(
	t *T, max ...int,
) (
	chan *lines.Events,
	func(lines.Componenter) *lines.Events,
) {
	chn := make(chan *lines.Events)
	return chn, func(c lines.Componenter) *lines.Events {
		ee, _ := lines.Test(t.GoT(), c, max...)
		chn <- ee
		return ee
	}
}

func (s *Gounit) Listens_to_events_if_not_fatale(t *T) {
	events, linesMock := mockLinesNew(t)
	go New(func(i ...interface{}) {
		t.Fatalf("unexpected error: %s", fmt.Sprint(i...))
	}, &watcherMock{}, linesMock)
	ee := <-events
	defer ee.QuitListening()
	t.Within(&TimeStepper{}, func() bool {
		return ee.IsListening()
	})
}

func TestGounit(t *testing.T) {
	t.Parallel()
	Run(&Gounit{}, t)
}
