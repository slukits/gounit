// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"time"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

// linesMask defines optional line-flags.
type linesMask map[uint]view.LineMask

// lines defines a reports lines content.
type rprLines []string

// report is the simplest implementation of view.Reporter.
type report struct {
	typ     reportType
	flags   view.RprtMask
	ll      rprLines
	llMasks linesMask
	lst     func(int)
}

// Flags control the processing of the report in the view.  E.g. if
// clearing is set all unused lines after an report update are cleared.
func (r *report) Flags() view.RprtMask { return r.flags }

// Type returns a report's type to determine a report transition after a
// user input; e.g. folding/un-folding suite-tests.
func (r *report) Type() reportType { return r.typ }

// For expects the view's reporting component and a callback to which
// the updated lines can be provided to.
func (r *report) For(_ lines.Componenter, line func(uint, string)) {
	for idx, content := range r.ll {
		line(uint(idx), content)
	}
}

// Mask returns for given index special formatting directives.
func (r *report) LineMask(idx uint) view.LineMask {
	if r.llMasks == nil {
		return view.ZeroLineMod
	}
	return r.llMasks[idx]
}

// Listener returns the callback which is informed about user selections
// of lines by providing the index of the selected line.
func (r *report) Listener() func(int) { return r.lst }

// setListener is part of the reporter-implementation.
func (r *report) setListener(l func(int)) {
	r.lst = l
}

func reportTestingPackage(p *pkg) []interface{} {
	ll, llMask := rprLines{}, linesMask{}
	if p.tp.LenSuites() == 0 {
		ll, llMask = reportGoTestsOnly(p, ll, llMask)
		return []interface{}{&report{
			flags:   view.RpClearing,
			ll:      ll,
			llMasks: llMask,
		}}
	}
	return reportMixedTests(p)
}

type suiteInfo struct {
	ttN, ffN int
	dr       time.Duration
}

func reportStatus(pp pkgs) *view.Statuser {
	// count suites, tests and failed tests
	ssLen, ttLen, ffLen := 0, 0, 0
	for _, p := range pp {
		ssLen += p.tp.LenSuites()
		p.tp.ForTest(func(t *model.Test) {
			n := p.Results.OfTest(t).Len()
			if n == 1 {
				ttLen++
				return
			}
			ssLen++
			ttLen += n
			ffLen += p.Results.OfTest(t).LenFailed()
		})
		p.tp.ForSuite(func(ts *model.TestSuite) {
			rr := p.OfSuite(ts)
			ttLen += rr.Len()
			ffLen += rr.LenFailed()
		})
	}
	return &view.Statuser{
		Packages: len(pp),
		Suites:   ssLen,
		Tests:    ttLen,
		Failed:   ffLen,
	}
}

func withFoldInfo(content string, tr *model.TestResult) string {
	return fmt.Sprintf("%s%s%d/%d %s",
		content, lines.LineFiller, tr.Len(), tr.LenFailed(),
		time.Duration(tr.End.Sub(tr.Start)).Round(1*time.Millisecond))
}
