// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/lines"
)

// Statuser instance passed to the provided status bar updater at
// initialization time
type Statuser struct {
	// Str is a status-bar string superseding all other status
	// information.
	Str string

	// Packages is the number of packages
	Packages int

	// Suites is the number of suites
	Suites int

	// Tests is the number of tests
	Tests int

	// Failed is the number of failed tests
	Failed int

	// Files is the number of code files
	Files int

	// TestFiles is the number of test-files
	TestFiles int

	// Lines is the number of code lines
	Lines int

	// TestLines is the number of test-code lines
	TestLines int

	// DocLines is the number of documentation lines
	DocLines int
}

type statusBar struct {
	lines.Component
	// np packages count
	np int
	// ns suites count
	ns int
	// nt tests count
	nt int
	// nf failed tests count
	nf int
	// ns source files count
	nsr int
	// nst source test files count
	nst int
	// nc code lines count
	nc int
	// nct code test lines count
	nct int
	// nd documentation lines count
	nd int
}

func (sb *statusBar) OnInit(e *lines.Env) {
	sb.Dim().SetHeight(2)
	fmt.Fprint(e.BG(sb.bg()).Filled().FG(sb.fg()).LL(1), sb.str())
}

func (sb *statusBar) OnUpdate(e *lines.Env) {
	// type save because message bar update only allows string
	s, _ := e.Evt.(*lines.UpdateEvent).Data.(Statuser)
	if s.Str != "" {
		fmt.Fprint(e.BG(sb.bg()).Filled().FG(sb.fg()).LL(1), s.Str)
		return
	}
	sb.np = s.Packages
	sb.ns = s.Suites
	sb.nt = s.Tests
	sb.nf = s.Failed
	sb.nsr = s.Files
	sb.nst = s.TestFiles
	sb.nc = s.Lines
	sb.nct = s.TestLines
	sb.nd = s.DocLines
	fmt.Fprint(e.BG(sb.bg()).Filled().FG(sb.fg()).LL(1), sb.str())
}

const dfltStatus = "pkgs/suites: %d/%d; tests: %d/%d"

func (sb *statusBar) str() string {
	pt := fmt.Sprintf(dfltStatus, sb.np, sb.ns, sb.nt, sb.nf)
	return pt
}

func (sb *statusBar) bg() tcell.Color {
	if sb.nf > 0 {
		return tcell.ColorRed
	}
	return tcell.ColorGreen
}

func (sb *statusBar) fg() tcell.Color {
	if sb.nf > 0 {
		return tcell.ColorWhite
	}
	return tcell.ColorBlack
}
