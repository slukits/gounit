// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"errors"
	"fmt"
	"testing"
	"time"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/lines"
)

// Gounit tests the behavior of Controller.New which is identical with
// the behavior of main.
type Gounit struct{ Suite }

func (s *Gounit) SetUp(t *T) { t.Parallel() }

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

func (s *Gounit) Fails_if_watching_fails(t *T) {
	mck := &watcherMock{watch: func() (<-chan *model.PackagesDiff, uint64, error) {
		return nil, 0, errors.New("mock-err")
	}}
	fatale := false

	New(func(_ ...interface{}) { fatale = true }, mck, nil)

	t.True(fatale)
}

type linesTest struct {
	ee *lines.Events
	tt *lines.Testing
}

func mockLinesNew(
	t *T, max ...int,
) (
	func(lines.Componenter) *lines.Events,
	func() (*lines.Events, *lines.Testing),
) {
	chn := make(chan struct{})
	var (
		ee *lines.Events
		tt *lines.Testing
	)
	return func(c lines.Componenter) *lines.Events {
			ee, tt = lines.Test(t.GoT(), c, max...)
			close(chn)
			return ee
		},
		func() (*lines.Events, *lines.Testing) {
			select {
			case <-t.Timeout(100 * time.Millisecond):
				t.Fatal("test: gounit: timeout: lines-initialization")
			case <-chn:
			}
			return ee, tt
		}
}

func (s *Gounit) Listens_to_events_if_not_fatale(t *T) {
	linesMock, linesTesting := mockLinesNew(t)
	go New(func(i ...interface{}) {
		t.Fatalf("unexpected error: %s", fmt.Sprint(i...))
	}, &watcherMock{}, linesMock)
	ee, _ := linesTesting()
	defer ee.QuitListening()
	t.Within(&TimeStepper{}, func() bool {
		return ee.IsListening()
	})
}

func (s *Gounit) Waits_for_testing_packages_if_there_are_none(t *T) {

}

func TestGounit(t *testing.T) {
	t.Parallel()
	Run(&Gounit{}, t)
}
