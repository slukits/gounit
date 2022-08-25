// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"os"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

// Gounit tests the behavior of Controller.New which is identical with
// the behavior of main.  Since we manipulate for each test the global
// variable "os.Args" we can neither run this suite nor it's tests in
// parallel.
type Gounit struct {
	Suite
	osArgs []string
}

func (s *Gounit) Init(t *S) { s.osArgs = os.Args }

func (s *Gounit) SetUp(t *T) { os.Args = s.osArgs }

func (s *Gounit) Fatales_not_implement_if_p_argument(t *T) {
	// TODO: needs removing if implemented
	args := []string{"p", "--p", "-p"}
	for _, a := range args {
		os.Args = []string{s.osArgs[0], a}
		New(func(i ...interface{}) {
			t.Contains(fmt.Sprint(i...), NotImplemented)
		}, lines.New)
	}
}

func (s *Gounit) Fatales_not_supported_if_other_arg_than_p(t *T) {
	os.Args = []string{s.osArgs[0], "42"}
	New(func(i ...interface{}) {
		t.Contains(fmt.Sprint(i...), fmt.Sprintf(NotSupported, "42"))
	}, lines.New)
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
	os.Args = []string{s.osArgs[0]}
	events, newMock := mockLinesNew(t)
	go New(func(i ...interface{}) {
		t.Fatalf("unexpected error: %s", fmt.Sprint(i...))
	}, newMock)
	ee := <-events
	defer ee.QuitListening()
	t.Within(&TimeStepper{}, func() bool {
		return ee.IsListening()
	})
}

func (s *Gounit) Finalize(t *S) { os.Args = s.osArgs }

func TestGounit(t *testing.T) {
	Run(&Gounit{}, t)
}
