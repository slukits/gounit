// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"testing"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type Reporting struct {
	Suite
	Fixtures
}

func (s *Reporting) SetUp(t *T) { t.Parallel() }

func (s *Reporting) TearDown(t *T) {
	quit := s.Get(t)
	if quit == nil {
		return
	}
	quit.(func())()
}

func (s *Reporting) fx(t *T) (*lines.Events, *lines.Testing, *viewFX) {
	return fx(t, s)
}

func (s *Reporting) Component_is_initially_focused(t *T) {
	ee, _, fx := s.fx(t)

	ee.Update(fx, nil, func(e *lines.Env) {
		t.Eq(fx.Report, e.Focused())
	})
}

func (s *Reporting) Component_is_scrollable(t *T) {
	ee, _, fx := s.fx(t)

	ee.Update(fx.Report, nil, func(e *lines.Env) {
		t.True(fx.Report.FF.Has(lines.Scrollable))
	})
}

func (s *Reporting) Component_s_lines_are_selectable(t *T) {
	ee, tt, fx := s.fx(t)

	ee.Update(fx.Report, nil, func(e *lines.Env) {
		t.True(fx.Report.FF.Has(lines.LinesSelectable))
		fmt.Fprint(e, "first\nsecond")
	})
	tt.FireKey(tcell.KeyDown)
	ee.Update(fx.Report, nil, func(e *lines.Env) {
		t.Eq(0, fx.Report.Focus.Current())
	})
}

func (s *Reporting) Component_s_test_lines_are_not_selectable(t *T) {
	ee, tt, fx := s.fx(t)
	fx.updateReporting(&reporterFX{
		ll: []string{"pkg 1", "test 1", "test 2", "suite 1"},
		mm: map[uint]LineMask{
			0: PackageLine,
			1: TestLine, 2: TestLine,
			3: SuiteLine,
		},
	})

	tt.FireKey(tcell.KeyDown)
	ee.Update(fx.Report, nil, func(e *lines.Env) {
		t.Eq(0, fx.Report.Focus.Current())
	})

	tt.FireKey(tcell.KeyDown)
	ee.Update(fx.Report, nil, func(e *lines.Env) {
		t.Eq(3, fx.Report.Focus.Current())
	})
}

func TestReporting(t *testing.T) {
	t.Parallel()
	Run(&Reporting{}, t)
}
