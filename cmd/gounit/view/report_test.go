// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type Reporting struct {
	Suite
}

func (s *Reporting) SetUp(t *T) { t.Parallel() }

func (s *Reporting) fx(t *T) *Fixture {
	return NewFixture(t, 0, nil)
}

func (s *Reporting) Component_is_initially_focused(t *T) {
	tt := s.fx(t)

	tt.Lines.Update(tt.Root(), nil, func(e *lines.Env) {
		t.Eq(tt.ReportCmp, e.Focused())
	})
}

func (s *Reporting) Component_is_scrollable(t *T) {
	tt := s.fx(t)

	tt.Lines.Update(tt.ReportCmp, nil, func(e *lines.Env) {
		t.True(tt.ReportCmp.FF.Has(lines.Scrollable))
	})
}

func (s *Reporting) Component_s_lines_are_selectable(t *T) {
	tt := s.fx(t)

	tt.Lines.Update(tt.ReportCmp, nil, func(e *lines.Env) {
		t.True(tt.ReportCmp.FF.Has(lines.LinesSelectable))
		fmt.Fprint(e, "first\nsecond")
	})
	tt.FireKey(lines.Down)
	tt.Lines.Update(tt.ReportCmp, nil, func(e *lines.Env) {
		t.Eq(0, tt.ReportCmp.LL.Focus.Current())
	})
}

func (s *Reporting) Component_s_test_lines_are_not_selectable(t *T) {
	tt := s.fx(t)
	tt.UpdateReporting(&reporterFX{
		ll: []string{"pkg 1", "test 1", "test 2", "suite 1"},
		mm: map[uint]LineMask{
			0: PackageLine,
			1: TestLine, 2: TestLine,
			3: GoSuiteLine,
		},
		listener: tt.defaultReportListener,
	})

	tt.FireKey(lines.Down)
	tt.Lines.Update(tt.ReportCmp, nil, func(e *lines.Env) {
		t.Eq(0, tt.ReportCmp.LL.Focus.Current())
	})

	tt.FireKey(lines.Down)
	tt.Lines.Update(tt.ReportCmp, nil, func(e *lines.Env) {
		t.Eq(3, tt.ReportCmp.LL.Focus.Current())
	})

	tt.FireComponentClick(tt.ReportCmp, 0, 3)
	t.Eq(3, tt.ReportedLine)
	tt.FireComponentClick(tt.ReportCmp, 0, 2)
	tt.Lines.Update(tt.ReportCmp, nil, func(e *lines.Env) {
		t.Eq(3, tt.ReportCmp.LL.Focus.Current())
	})
	t.Eq(3, tt.ReportedLine)
}

func (s *Reporting) Component_s_selectable_lines_are_underlined(t *T) {
	tt := s.fx(t)
	tt.UpdateReporting(&reporterFX{
		content:  "selectable line",
		mm:       map[uint]LineMask{0: PackageLine},
		listener: tt.defaultReportListener,
	})

	line := tt.CellsOf(tt.ReportCmp)[0]
	for i, r := range line.String() {
		if r == ' ' {
			continue
		}
		t.FatalIfNot(t.True(line.HasAA(i, lines.Underline)))
	}
}

func TestReporting(t *testing.T) {
	t.Parallel()
	Run(&Reporting{}, t)
}
