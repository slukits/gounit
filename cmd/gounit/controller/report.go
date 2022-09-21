// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"sort"
	"strings"
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
	flags   view.RprtMask
	ll      rprLines
	llMasks linesMask
	lst     func(int)
}

// Flags control the processing of the report in the view.  E.g. if
// clearing is set all unused lines after an report update are cleared.
func (r *report) Flags() view.RprtMask { return r.flags }

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

// reportPackages reports all packages of watched source directory
// folded.
func reportPackages(pp pkgs, lst func(int)) *report {
	ll, llMask := rprLines{}, linesMask{}
	_pp := []*pkg{}
	for _, p := range pp {
		_pp = append(_pp, p)
	}
	sort.Slice(_pp, func(i, j int) bool {
		return _pp[i].ID() < _pp[j].ID()
	})
	for _, p := range _pp {
		ll, llMask = reportPackageLine(
			p, view.PackageFoldedLine, ll, llMask)
	}
	return &report{
		flags:   view.RpClearing,
		ll:      ll,
		llMasks: llMask,
		lst:     lst,
	}
}

func reportFailed(st *state, lst func(int)) *report {
	ll, llMask := rprLines{}, linesMask{}
	for _, p := range st.ee[:len(st.ee)-1] {
		ll, llMask = reportPackageLine(
			p, view.PackageFoldedLine, ll, llMask)
	}
	ll, llMask = reportFailedPkg(st.ee[len(st.ee)-1], ll, llMask)
	return &report{
		flags:   view.RpClearing,
		ll:      ll,
		llMasks: llMask,
		lst:     lst,
	}
}

func reportFailedPkg(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
	if p.HasErr() {
		for _, s := range strings.Split(p.Err(), "\n") {
			ll = append(ll, s)
			llMask[uint(len(ll)-1)] = view.OutputLine
		}
		return ll, llMask
	}

	ll, llMask, goTestsFailed := reportFailedPkgGoTests(p, ll, llMask)

	if goTestsFailed {
		return reportFailedPkgSuiteHeader(p, ll, llMask)
	}
	return reportFailedPkgSuites(p, ll, llMask)
}

// reportFailedPkgGoTests reports (potentially) failed go tests.  If
// there are no go tests nothing is done; if there are go-tests which
// are all passing only the "go-tests"-line is reported.  In the later
// two cases false is returned; true otherwise.
func reportFailedPkgGoTests(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask, bool) {
	if p.LenTests() == 0 {
		return ll, llMask, false
	}
	n, f, d, without, with := goSplitTests(p)
	if f == 0 {
		ll, llMask = reportGoTestsLine(
			n, f, d, view.GoTestsFoldedLine, ll, llMask)
		return ll, llMask, false
	}
	ll, llMask = reportGoTestsLine(n, f, d, view.GoTestsLine, ll, llMask)
	if without.haveFailed(p) {
		ll = append(ll, blankLine)
		for _, t := range without {
			ll, llMask = reportTestLine(p.OfTest(t), indent, ll, llMask)
		}
		ll = append(ll, blankLine)
		ll, llMask = reportGoSuitesFolded(p, with, ll, llMask)
		return ll, llMask, true
	}
	var tt []*model.Test
	for _, t := range with {
		if p.OfTest(t).Passed {
			continue
		}
		tt = append(tt, t)
	}
	for _, t := range tt[:len(tt)-1] {
		ll, llMask = reportGoSuiteLine(
			p.OfTest(t), view.GoSuiteFoldedLine, ll, llMask)
	}
	reportGoSuite(p.OfTest(tt[len(tt)-1]), ll, llMask)
	return ll, llMask, true
}

func reportGoSuitesFolded(
	p *pkg, with goTests, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	for _, t := range with {
		ll, llMask = reportGoSuiteLine(
			p.OfTest(t), view.GoSuiteFoldedLine, ll, llMask)
	}
	return ll, llMask
}

func reportGoSuite(
	r *model.TestResult, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll, llMask = reportGoSuiteLine(r, view.GoSuiteLine, ll, llMask)
	r.ForOrdered(func(r *model.SubResult) {
		ll, llMask = reportSubTestLine(r, indent+indent, ll, llMask)
	})
	return ll, llMask
}

func reportFailedPkgSuiteHeader(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	p.ForSuite(func(s *model.TestSuite) {
		if p.OfSuite(s).Passed {
			return
		}
		ll, llMask = reportSuiteLine(
			p, s, view.SuiteFoldedLine, ll, llMask)
	})
	return ll, llMask
}

func reportFailedPkgSuites(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	ss := []*model.TestSuite{}
	p.ForSuite(func(s *model.TestSuite) {
		if p.OfSuite(s).Passed {
			return
		}
		ss = append(ss, s)
	})
	if len(ss) == 0 {
		return ll, llMask
	}

	for _, s := range ss[:len(ss)-1] {
		ll, llMask = reportSuiteLine(
			p, s, view.SuiteFoldedLine, ll, llMask)
	}
	ll, llMask = reportSuite(p, ss[len(ss)-1], ll, llMask)
	return ll, llMask
}

func reportSuite(
	p *pkg, s *model.TestSuite, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	ll, llMask = reportSuiteLine(p, s, view.SuiteLine, ll, llMask)
	rr := p.OfSuite(s)
	for i, out := range rr.InitOut {
		if i == 0 {
			ll = append(ll, indent+"init-log:")
		}
		ll = append(ll, indent+indent+strings.TrimSpace(out))
	}
	if len(rr.InitOut) > 0 {
		ll = append(ll, blankLine)
	}
	s.ForTest(func(t *model.Test) {
		ll, llMask = reportSubTestLine(rr.OfTest(t), indent, ll, llMask)
	})
	for i, out := range rr.FinalizeOut {
		if i == 0 {
			ll = append(ll, blankLine, indent+"finalize-log:")
		}
		ll = append(ll, indent+indent+strings.TrimSpace(out))
	}
	return ll, llMask
}

func withFoldInfo(content string, tr *model.TestResult) string {
	return fmt.Sprintf("%s%s%d/%d %s",
		content, lines.LineFiller, tr.Len(), tr.LenFailed(),
		time.Duration(tr.End.Sub(tr.Start)).Round(1*time.Millisecond))
}

func reportPackageLine(
	p *pkg, lm view.LineMask, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	n, f, d := p.info()
	ll = append(ll, fmt.Sprintf("%s%s%d/%d %s",
		p.ID(), lines.LineFiller, n, f, d.Round(1*time.Millisecond)))
	idx := uint(len(ll) - 1)
	llMask[idx] = lm
	if f > 0 {
		llMask[idx] |= view.Failed
	}
	return ll, llMask
}

func reportGoTestsLine(
	n, f int, d time.Duration, lm view.LineMask,
	ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	content := fmt.Sprintf("go-tests%s%d/%d %s",
		lines.LineFiller, n, f, d.Round(1*time.Millisecond))
	ll = append(ll, content)
	idx := uint(len(ll) - 1)
	llMask[idx] = lm
	if f > 0 {
		llMask[idx] |= view.Failed
	}
	return ll, llMask
}

func reportGoSuiteLine(
	r *model.TestResult, lm view.LineMask, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	content := indent + r.String()
	if lm == view.GoSuiteFoldedLine {
		content = fmt.Sprintf("%s%s%d/%d %s",
			content, lines.LineFiller, r.Len(), r.LenFailed(),
			r.End.Sub(r.Start).Round(1*time.Millisecond))
	}
	ll = append(ll, content)
	idx := uint(len(ll) - 1)
	llMask[idx] = lm
	if !r.Passed {
		llMask[idx] |= view.Failed
	}
	return ll, llMask
}

func reportSuiteLine(
	p *pkg, s *model.TestSuite, lm view.LineMask,
	ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	content := s.String()
	r := p.OfSuite(s)
	if lm == view.SuiteFoldedLine {
		content = fmt.Sprintf("%s%s%d/%d %s", content,
			lines.LineFiller, r.Len(), r.LenFailed(),
			r.End.Sub(r.Start).Round(1*time.Millisecond))
	}
	ll = append(ll, content)
	idx := uint(len(ll) - 1)
	llMask[idx] = lm
	if r.LenFailed() > 0 {
		llMask[idx] |= view.Failed
	}
	return ll, llMask
}

func reportTestLine(
	r *model.TestResult, i string, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll = append(ll, fmt.Sprintf("%s%s%s",
		indent+r.String(),
		lines.LineFiller,
		r.End.Sub(r.Start).Round(1*time.Millisecond)))
	llMask[uint(len(ll)-1)] = view.TestLine
	for _, out := range r.Output {
		ll = append(ll, i+indent+strings.TrimSpace(out))
		llMask[uint(len(ll)-1)] = view.OutputLine
	}
	return ll, llMask
}

func reportSubTestLine(
	r *model.SubResult, i string, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll = append(ll, fmt.Sprintf("%s%s%s",
		i+r.String(),
		lines.LineFiller,
		r.End.Sub(r.Start).Round(1*time.Millisecond)))
	llMask[uint(len(ll)-1)] = view.TestLine
	for _, out := range r.Output {
		ll = append(ll, i+indent+strings.TrimSpace(out))
		llMask[uint(len(ll)-1)] = view.OutputLine
	}
	return ll, llMask
}

func reportStatus(pp pkgs) *view.Statuser {
	// count suites, tests and failed tests
	ssLen, ttLen, ffLen := 0, 0, 0
	for _, p := range pp {
		ssLen += p.LenSuites()
		p.ForTest(func(t *model.Test) {
			n := p.Results.OfTest(t).Len()
			if n == 1 {
				ttLen++
				return
			}
			ssLen++
			ttLen += n
			ffLen += p.Results.OfTest(t).LenFailed()
		})
		p.ForSuite(func(ts *model.TestSuite) {
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
