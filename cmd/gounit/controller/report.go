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

// newReport calculates from given state and requested report type a
// report and returns it.  The index is needed for the use case that a
// folded package/suite selected by the user needs to be determined.  In
// case given report type and state are inconsistent newReport falls
// back reporting all packages folded.  The later may happen if
// user-input and watch-update happen at the same time.
func newReport(st *state, t reportType, idx int) *report {
	if len(st.pp) == 0 {
		return emptyReport
	}
	if t == rprPackages {
		ll, llMask := reportPackages(st.pp)
		st.latestPkg = ""
		return &report{ll: ll, llMasks: llMask}
	}
	if t == rprPackage {
		ln, ok := findReportLine(st.view[0].(*report), idx,
			view.PackageFoldedLine|view.PackageLine)
		if !ok {
			ll, llMask := reportPackages(st.pp)
			return &report{ll: ll, llMasks: llMask}
		}
		st.latestPkg = ln
	}
	p := st.ensureLatestPackage()
	ll, llMask := rprLines{}, linesMask{}
	if len(st.ee) > 0 {
		ll, llMask = reportFailedPkgsBut(st, ll, llMask)
	}
	if st.ee[st.latestPkg] {
		ll, llMask := reportFailedPkg(st, t, idx, ll, llMask)
		return &report{ll: ll, llMasks: llMask}
	}
	switch p.LenSuites() {
	case 0:
		switch t {
		case rprGoSuite:
			gs := findGoSuite(st, p, idx)
			if gs == nil {
				ll, llMask := reportPackages(st.pp)
				return &report{ll: ll, llMasks: llMask}
			}
			st.lastSuite = "go-tests:" + gs.Name()
			ll, llMask = reportGoOnlySuite(p, gs, ll, llMask)
		default:
			ll, llMask = reportGoOnlyPkg(p, ll, llMask)
		}
	default:
		ll, llMask = reportMixedPkg(st, t, idx, p, ll, llMask)
	}
	return &report{ll: ll, llMasks: llMask}
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

func reportFailedPkgsBut(
	st *state, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ee := []string{}
	for k := range st.ee {
		ee = append(ee, k)
	}
	sort.Slice(ee, func(i, j int) bool {
		return ee[i] < ee[j]
	})
	for _, pID := range ee {
		if pID == st.latestPkg {
			continue
		}
		ll, llMask = reportPackageLine(
			st.pp[pID], view.PackageFoldedLine, ll, llMask)
	}
	return ll, llMask
}

// reportMixedPkg reports mixed package according to given state and
// report-type.  reportMixedPkg falls back to report the package
// "folded" iff the report-type cannot be reported with given state.
// This may happen if a user-input and watch-update happens at the same
// time.
func reportMixedPkg(
	st *state, t reportType, idx int, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	switch t {
	case rprGoTests, rprGoSuiteFolded:
		if p.LenTests() == 0 {
			return reportMixedFolded(p, ll, llMask)
		}
		st.lastSuite = "go-tests"
		return reportMixedGoTests(p, ll, llMask)
	case rprSuite:
		suite := findSuite(st, p, idx)
		if suite == nil {
			return reportMixedFolded(p, ll, llMask)
		}
		st.lastSuite = suite.Name()
		return reportMixedSuite(suite, p, ll, llMask)
	case rprGoSuite:
		gSuite := findGoSuite(st, p, idx)
		if gSuite == nil {
			return reportMixedFolded(p, ll, llMask)
		}
		st.lastSuite = "go-tests:" + gSuite.Name()
		return reportMixedGoSuite(gSuite, p, ll, llMask)
	case rprMixedFolded:
		return reportMixedFolded(p, ll, llMask)
	case rprDefault, rprPackage:
		var suite *model.TestSuite
		if st.lastSuite != "" {
			if strings.HasPrefix(st.lastSuite, "go-tests") {
				return reportLockedGoSuite(st, p, ll, llMask)
			}
			suite = p.Suite(st.lastSuite)
		}
		if suite == nil {
			st.lastSuite = ""
			suite = p.LastSuite()
		}
		if suite != nil {
			return reportMixedSuite(suite, p, ll, llMask)
		}
		return reportMixedFolded(p, ll, llMask)
	}
	return ll, llMask
}

func findSuite(st *state, p *pkg, idx int) *model.TestSuite {
	ln, ok := findReportLine(st.view[0].(*report), idx,
		view.SuiteFoldedLine|view.SuiteLine)
	if !ok {
		return nil
	}

	var suite *model.TestSuite
	p.ForSuite(func(ts *model.TestSuite) {
		if suite != nil {
			return
		}
		if ln == ts.String() {
			suite = ts
		}
	})

	return suite
}

func reportLockedGoSuite(
	st *state, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	if p.LenTests() == 0 {
		return reportMixedFolded(p, ll, llMask)
	}
	if st.lastSuite == "go-tests" {
		return reportMixedGoTests(p, ll, llMask)
	}
	goSuiteName := strings.Split(st.lastSuite, ":")[1]
	var goSuite *model.Test
	p.ForTest(func(t *model.Test) {
		if goSuite != nil {
			return
		}
		if t.Name() != goSuiteName {
			return
		}
		goSuite = t
	})
	if goSuite != nil {
		return reportMixedGoSuite(goSuite, p, ll, llMask)
	}
	return reportMixedFolded(p, ll, llMask)
}

func findGoSuite(st *state, p *pkg, idx int) *model.Test {
	var goSuite *model.Test
	ln, ok := findReportLine(st.view[0].(*report), idx,
		view.GoSuiteFoldedLine|view.GoSuiteLine)
	if !ok {
		return nil
	}
	p.ForTest(func(t *model.Test) {
		if goSuite != nil {
			return
		}
		if ln == t.String() {
			goSuite = t
		}
	})
	return goSuite
}

func findReportLine(r *report, idx int, lm view.LineMask) (string, bool) {
	if idx >= len(r.ll) || r.LineMask(uint(idx))&lm == 0 {
		return "", false
	}
	return strings.TrimSpace(strings.Split(
		r.ll[idx], lines.LineFiller)[0]), true
}

// reportPackages reports all packages of watched source directory
// folded.
func reportPackages(pp pkgs) (rprLines, linesMask) {
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
	return ll, llMask
}

func reportFailedPkg(
	st *state, rt reportType, idx int, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	p := st.pp[st.latestPkg]
	if p.HasErr() {
		ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
		ll = append(ll, blankLine)
		ll, llMask = reportPkgErr(p, ll, llMask)
		return ll, llMask
	}
	var gs *model.Test
	if rt == rprGoSuite {
		gs = findGoSuite(st, p, idx)
		if gs == nil {
			ll, llMask = reportPackages(st.pp)
			return ll, llMask
		}
	}

	if p.LenSuites() == 0 {
		switch rt {
		case rprGoSuite:
			ll, llMask = reportGoOnlySuite(p, gs, ll, llMask)
			return ll, llMask
		default:
			ll, llMask = reportGoOnlyPkg(p, ll, llMask)
			return ll, llMask
		}
	}

	ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
	ll = append(ll, blankLine)
	ll, llMask, goTestsFailed := reportFailedPkgGoTests(p, ll, llMask, gs)

	if goTestsFailed && rt != rprSuite {
		return reportFailedPkgSuitesHeader(p, ll, llMask)
	}

	var suite *model.TestSuite
	if rt == rprSuite {
		suite = findSuite(st, p, idx)
	}
	return reportFailedPkgSuites(p, ll, llMask, suite)
}

var reHex = regexp.MustCompile(`\s*[+]0x[0-9a-z]*\s*`)

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
			if reHex.MatchString(strings.TrimSpace(s[n:])) {
				continue
			}
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
	p *pkg, ll rprLines, llMask linesMask, gs *model.Test,
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
			if gs != nil && p.OfTest(t).Passed {
				continue
			}
			ll, llMask = reportTestLine(p, p.OfTest(t), indent, ll, llMask)
		}
		ll = append(ll, blankLine)
		if gs == nil {
			ll, llMask = reportGoSuitesFolded(p, with, ll, llMask)
			return ll, llMask, true
		}
	}
	var tt []*model.Test
	for _, t := range with {
		if p.OfTest(t).Passed {
			continue
		}
		tt = append(tt, t)
	}
	if gs == nil {
		gs = tt[len(tt)-1]
	}
	for _, t := range tt {
		if t.Name() == gs.Name() {
			continue
		}
		ll, llMask = reportGoSuiteLine(
			p.OfTest(t), view.GoSuiteFoldedLine, indent, ll, llMask)
	}
	ll, llMask = reportGoSuite(p, p.OfTest(gs), ll, llMask)
	return ll, llMask, true
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
	p *pkg, ll rprLines, llMask linesMask, suite *model.TestSuite,
) (rprLines, linesMask) {

	ss := []*model.TestSuite{}
	p.ForSuite(func(s *model.TestSuite) {
		if p.OfSuite(s).Passed {
			return
		}
		ss = append(ss, s)
	})
	if len(ss) == 0 && suite == nil {
		return ll, llMask
	}
	sort.Slice(ss[:len(ss)-1], func(i, j int) bool {
		return ss[i].Name() < ss[j].Name()
	})

	if suite == nil {
		suite = ss[len(ss)-1]
	}
	for _, s := range ss {
		if s.Name() == suite.Name() {
			continue
		}
		ll, llMask = reportSuiteLine(
			p, s, view.SuiteFoldedLine, ll, llMask)
	}
	ll, llMask = reportSuite(p, suite, ll, llMask)
	return ll, llMask
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

var raceLineBreak = regexp.MustCompile(`^\s*Goroutine \d+|^\s*Read at` +
	`|^\s*Previous read at|^\s*Write at|^\s*Previous write at`)

func reportOutput(
	p *pkg, out []string, i string, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	if len(out) == 0 {
		return ll, llMask
	}
	var dataRace bool
	for _, s := range out {
		if strings.Contains(s, "WARNING: DATA RACE") {
			dataRace = true
		}
		if flLoc, n, ok := pkgFileLoc(p, s); ok {
			if dataRace {
				ll, llMask = reportOutputLine(
					outputWidth, flLoc, i+indent, ll, llMask)
			} else {
				ll, llMask = reportOutputLine(
					outputWidth, flLoc, i, ll, llMask)
			}
			if reHex.MatchString(strings.TrimSpace(s[n:])) {
				continue
			}
			ll, llMask = reportOutputLine(
				outputWidth, strings.TrimSpace(s[n:]), i+indent,
				ll, llMask)
			continue
		}
		if dataRace && raceLineBreak.MatchString(s) {
			ll = append(ll, blankLine)
			llMask[uint(len(ll)-1)] = view.OutputLine
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
var reFileLoc = regexp.MustCompile(`^.*?.go:[0-9]*[:]*`)

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

func newStatus(pp pkgs) *view.Statuser {
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
