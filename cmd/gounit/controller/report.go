// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
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

var emptyReport = &report{
	ll:    []string{initReport},
	flags: view.RpClearing,
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
	ee := []string{}
	for pID := range st.ee {
		ee = append(ee, pID)
	}
	sort.Slice(ee, func(i, j int) bool {
		return ee[i] < ee[j]
	})
	if !st.ee[st.latestPkg] {
		st.latestPkg = ee[len(ee)-1]
	}
	for _, pID := range ee {
		if pID == st.latestPkg {
			continue
		}
		ll, llMask = reportPackageLine(
			st.pp[pID], view.PackageFoldedLine, ll, llMask)
	}
	ll, llMask, ls := reportFailedPkg(st, ll, llMask)
	st.lastSuite = ls
	return &report{
		flags:   view.RpClearing,
		ll:      ll,
		llMasks: llMask,
		lst:     lst,
	}
}

func reportFailedPkg(
	st *state, ll rprLines, llMask linesMask,
) (rprLines, linesMask, string) {

	p := st.pp[st.latestPkg]
	if p.HasErr() {
		ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
		ll = append(ll, blankLine)
		ll, llMask = reportPkgErr(p, ll, llMask)
		return ll, llMask, ""
	}
	if p.LenSuites() == 0 {
		ll, llMask = reportGoOnlyPkg(p, ll, llMask)
		return ll, llMask, "go-tests"
	}

	ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
	ll = append(ll, blankLine)
	ll, llMask, goTestsFailed := reportFailedPkgGoTests(p, ll, llMask)

	if goTestsFailed != "" {
		ll, llMask = reportFailedPkgSuitesHeader(p, ll, llMask)
		return ll, llMask, goTestsFailed
	}

	return reportFailedPkgSuites(p, ll, llMask)
}

func reportPkgErr(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	for _, s := range strings.Split(p.Err(), "\n") {
		if strings.HasPrefix(s, "# ") {
			continue
		}
		s = strings.TrimPrefix(s, "./")
		if flLoc, n, ok := pkgFileLoc(p, s); ok {
			ll, llMask = reportOutputLine(
				outputWidth, flLoc, indent, ll, llMask)
			ll, llMask = reportOutputLine(
				outputWidth, strings.TrimSpace(s[n:]), indent+indent,
				ll, llMask)
			continue
		}
		ll, llMask = reportOutputLine(outputWidth, s, indent, ll, llMask)
	}
	return ll, llMask
}

// reportFailedPkgGoTests reports (potentially) failed go tests.  If
// there are no go tests nothing is done; if there are go-tests which
// are all passing only the "go-tests"-line is reported.  In the later
// two cases false is returned; true otherwise.
func reportFailedPkgGoTests(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask, string) {
	if p.LenTests() == 0 {
		return ll, llMask, ""
	}
	n, f, d, without, with := goSplitTests(p)
	if f == 0 {
		ll, llMask = reportGoTestsLine(
			n, f, d, view.GoTestsFoldedLine, ll, llMask)
		return ll, llMask, ""
	}
	ll, llMask = reportGoTestsLine(n, f, d, view.GoTestsLine, ll, llMask)
	if without.haveFailed(p) {
		ll = append(ll, blankLine)
		for _, t := range without {
			ll, llMask = reportTestLine(p, p.OfTest(t), indent, ll, llMask)
		}
		ll = append(ll, blankLine)
		ll, llMask = reportGoSuitesFolded(p, with, ll, llMask)
		return ll, llMask, "go-tests"
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
			p.OfTest(t), view.GoSuiteFoldedLine, indent, ll, llMask)
	}
	reportGoSuite(p, p.OfTest(tt[len(tt)-1]), ll, llMask)
	return ll, llMask, fmt.Sprintf(
		"go-tests:%s", tt[len(tt)-1].Name())
}

func reportGoSuitesFolded(
	p *pkg, with goTests, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	for _, t := range with {
		ll, llMask = reportGoSuiteLine(
			p.OfTest(t), view.GoSuiteFoldedLine, indent, ll, llMask)
	}
	return ll, llMask
}

func reportGoSuite(
	p *pkg, r *model.TestResult, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll, llMask = reportGoSuiteLine(r, view.GoSuiteLine, indent, ll, llMask)
	r.ForOrdered(func(r *model.SubResult) {
		ll, llMask = reportSubTestLine(p, r, indent+indent, ll, llMask)
	})
	return ll, llMask
}

func reportFailedPkgSuitesHeader(
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
	ll = append(ll, blankLine)
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Name() < ss[j].Name()
	})

	for _, s := range ss {
		ll, llMask = reportSuiteLine(
			p, s, view.SuiteFoldedLine, ll, llMask)
	}
	return ll, llMask
}

func reportFailedPkgSuites(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask, string) {

	ss := []*model.TestSuite{}
	p.ForSuite(func(s *model.TestSuite) {
		if p.OfSuite(s).Passed {
			return
		}
		ss = append(ss, s)
	})
	if len(ss) == 0 {
		return ll, llMask, ""
	}
	sort.Slice(ss[:len(ss)-1], func(i, j int) bool {
		return ss[i].Name() < ss[j].Name()
	})

	for _, s := range ss[:len(ss)-1] {
		ll, llMask = reportSuiteLine(
			p, s, view.SuiteFoldedLine, ll, llMask)
	}
	ll, llMask = reportSuite(p, ss[len(ss)-1], ll, llMask)
	return ll, llMask, ss[len(ss)-1].Name()
}

func reportSuite(
	p *pkg, s *model.TestSuite, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	ll, llMask = reportSuiteLine(p, s, view.SuiteLine, ll, llMask)
	rr := p.OfSuite(s)
	if len(rr.InitOut) > 0 {
		ll = append(ll, indent+"init-log:")
		llMask[uint(len(ll)-1)] = view.OutputLine
		ll, llMask = reportOutput(
			p, rr.InitOut, indent, ll, llMask)
		ll = append(ll, blankLine)
	}
	s.ForTest(func(t *model.Test) {
		ll, llMask = reportSubTestLine(p, rr.OfTest(t), indent, ll, llMask)
	})
	if len(rr.FinalizeOut) > 0 {
		ll = append(ll, indent+"finalize-log:")
		llMask[uint(len(ll)-1)] = view.OutputLine
		ll, llMask = reportOutput(
			p, rr.FinalizeOut, indent, ll, llMask)
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

	if p.HasErr() {
		ll = append(ll, p.ID())
		idx := uint(len(ll) - 1)
		llMask[idx] = lm | view.Failed
		return ll, llMask
	}
	n, f, _, d := p.info()
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

	content := "go-tests"
	if lm == view.GoTestsFoldedLine {
		content = fmt.Sprintf("%s%s%d/%d %s", content,
			lines.LineFiller, n, f, d.Round(1*time.Millisecond))
	}
	ll = append(ll, content)
	idx := uint(len(ll) - 1)
	llMask[idx] = lm
	if f > 0 {
		llMask[idx] |= view.Failed
	}
	return ll, llMask
}

func reportGoSuiteLine(
	r *model.TestResult, lm view.LineMask, i string,
	ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	content := i + r.String()
	if lm == view.GoSuiteFoldedLine {
		content = fmt.Sprintf("%s%s%d/%d",
			content, lines.LineFiller, r.Len(), r.LenFailed())
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
		content = fmt.Sprintf("%s%s%d/%d", content,
			lines.LineFiller, r.Len(), r.LenFailed())
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
	p *pkg, r *model.TestResult, i string, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll = append(ll, fmt.Sprintf("%s%s",
		i+r.String(),
		lines.LineFiller))
	idx := uint(len(ll) - 1)
	llMask[idx] = view.TestLine
	if !r.Passed {
		llMask[idx] = view.Failed
	}
	return reportOutput(p, r.Output, i+indent, ll, llMask)
}

func reportSubTestLine(
	p *pkg, r *model.SubResult, i string, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll = append(ll, fmt.Sprintf("%s%s",
		i+r.String(),
		lines.LineFiller))
	idx := uint(len(ll) - 1)
	llMask[idx] = view.TestLine
	if !r.Passed {
		llMask[idx] = view.Failed
	}
	return reportOutput(p, r.Output, i+indent, ll, llMask)
}

const outputWidth = 68

func reportOutput(
	p *pkg, out []string, i string, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	if len(out) == 0 {
		return ll, llMask
	}
	for _, s := range out {
		if flLoc, n, ok := pkgFileLoc(p, s); ok {
			ll, llMask = reportOutputLine(
				outputWidth, flLoc, i, ll, llMask)
			ll, llMask = reportOutputLine(
				outputWidth, strings.TrimSpace(s[n:]), i+indent,
				ll, llMask)
			continue
		}
		ll, llMask = reportOutputLine(
			outputWidth, s, i+indent, ll, llMask)
	}
	return ll, llMask
}

func reportOutputLine(
	width int, out, i string, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	out = i + out
	if len(out) < width {
		ll = append(ll, out)
		llMask[uint(len(ll)-1)] = view.OutputLine
		return ll, llMask
	}

	subIndent := len(i + indent)
	oo, rest := []string{out[:width]}, out[width:]
	for len(rest)+subIndent >= width {
		oo = append(oo, strings.TrimSpace(rest[:width-subIndent]))
		rest = rest[width-subIndent:]
	}
	oo = append(oo, strings.TrimSpace(rest))
	ll = append(ll, oo[0])
	llMask[uint(len(ll)-1)] = view.OutputLine
	for _, s := range oo[1:] {
		ll = append(ll, i+indent+s)
		llMask[uint(len(ll)-1)] = view.OutputLine
	}

	return ll, llMask
}

var reFilePos = regexp.MustCompile(`^.*?.go:[0-9]*?:[0-9]*?:`)
var reFileLoc = regexp.MustCompile(`^.*?.go:[0-9]*`)

func pkgFileLoc(p *pkg, s string) (loc string, n int, ok bool) {
	flLoc := reFilePos.FindString(s)
	if flLoc == "" {
		flLoc = reFileLoc.FindString(s)
		if flLoc == "" {
			return "", 0, false
		}
	}
	idx := strings.Index(flLoc, ":")
	pkgPrefix := strings.TrimSuffix(p.Abs(), filepath.Dir(p.ID()))
	if strings.HasPrefix(flLoc, pkgPrefix) {
		return strings.TrimPrefix(flLoc, pkgPrefix), len(flLoc), true
	}
	_, err := os.Stat(filepath.Join(
		p.Abs(), filepath.Base(p.ID()), flLoc[:idx]))
	if err != nil {
		return "", 0, false
	}
	return filepath.Join(p.ID(), flLoc), len(flLoc), true
}

func reportStatus(pp pkgs) *view.Statuser {
	// count suites, tests and failed tests
	ssLen, ttLen, ffLen := 0, 0, 0
	for _, p := range pp {
		n, f, s, _ := p.info()
		ssLen += s
		ttLen += n
		ffLen += f
	}
	return &view.Statuser{
		Packages: len(pp),
		Suites:   ssLen,
		Tests:    ttLen,
		Failed:   ffLen,
	}
}
