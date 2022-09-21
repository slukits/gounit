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

func withFoldInfo(content string, tr *model.TestResult) string {
	return fmt.Sprintf("%s%s%d/%d %s",
		content, lines.LineFiller, tr.Len(), tr.LenFailed(),
		time.Duration(tr.End.Sub(tr.Start)).Round(1*time.Millisecond))
}

func packageLine(
	p *pkg, ll rprLines, llMask linesMask,
) (func(ll rprLines, llMask linesMask, n, f int) (rprLines, linesMask),
	rprLines, linesMask,
) {

	idx := uint(len(ll))
	ll = append(ll, blankLine, blankLine)

	return func(
		ll rprLines, llMask linesMask, n, f int,
	) (rprLines, linesMask) {

		ll[idx] = fmt.Sprintf("%s: %d/%d %v", p.ID(), n, f,
			p.Duration.Round(1*time.Millisecond))
		llMask[idx] = view.PackageLine
		return ll, llMask
	}, ll, llMask
}

func foldedSuiteLine(
	ts *model.TestSuite, r *model.TestResult,
	ll rprLines, llMask linesMask,
) (rprLines, linesMask, int, int) {
	n, f, d := r.Len(), r.LenFailed(), r.End.Sub(r.Start)
	content := fmt.Sprintf("%s%s%d/%d %s",
		ts.String(), lines.LineFiller, n, f,
		d.Round(1*time.Millisecond))
	ll = append(ll, content)
	idx := uint(len(ll) - 1)
	llMask[idx] = view.SuiteFoldedLine
	if f > 0 {
		llMask[idx] |= view.Failed
	}
	return ll, llMask, n, f
}

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
		n, f, d := p.info()
		ll = append(ll, fmt.Sprintf("%s%s%d/%d %s",
			p.ID(), lines.LineFiller, n, f, d.Round(1*time.Millisecond)))
		llMask[uint(len(ll)-1)] = view.PackageFoldedLine
	}
	return &report{
		flags:   view.RpClearing,
		ll:      ll,
		llMasks: llMask,
		lst:     lst,
	}
}

func reportResult(
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

func reportSubResult(
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
