// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"strings"
	"time"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

// indent of a reported test-name.
const indent = "  "

func reportMixedFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	pkgLine, ll, llMask := packageLine(p, ll, llMask)
	n, f := 0, 0
	if p.LenTests() > 0 {
		ll, llMask, n, f = reportGoTestsFolded(p, ll, llMask)
	}
	var _n, _f int
	ll, llMask, _n, _f = reportSuitesFolded(p, ll, llMask)
	n += _n
	f += _f
	ll, llMask = pkgLine(ll, llMask, n, f)
	return ll, llMask
}

func reportMixedGoTests(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	n, f, d := p.info()
	ll = append(ll, fmt.Sprintf("%s: %d/%d %v", p.ID(), n, f,
		d.Round(1*time.Millisecond)))
	llMask[uint(len(ll)-1)] = view.PackageLine
	ll = append(ll, blankLine)
	ll, llMask = reportGoTestsSubTestsFolded(p, ll, llMask)
	return ll, llMask
}

func reportMixedGoSuite(
	goSuite *model.Test, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	n, f, d := p.info()
	ll = append(ll, fmt.Sprintf("%s: %d/%d %v", p.ID(), n, f,
		d.Round(1*time.Millisecond)))
	llMask[uint(len(ll)-1)] = view.PackageLine
	ll = append(ll, blankLine)
	n, f, d, _, _ = goSplitTests(p)
	content := fmt.Sprintf("go-tests%s%d/%d %s",
		lines.LineFiller, n, f, d.Round(1*time.Millisecond))
	ll = append(ll, content)
	llMask[uint(len(ll)-1)] = view.GoTestsLine
	ll = append(ll, indent+goSuite.String())
	llMask[uint(len(ll)-1)] = view.GoSuiteLine
	tr := p.OfTest(goSuite)
	tr.ForOrdered(func(sr *model.SubResult) {
		ll, llMask = reportSubResult(sr, indent+indent, ll, llMask)
	})
	return ll, llMask
}

func reportMixedSuite(
	suite *model.TestSuite, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	n, f, d := p.info()
	ll = append(ll, fmt.Sprintf("%s: %d/%d %v", p.ID(), n, f,
		d.Round(1*time.Millisecond)))
	llMask[uint(len(ll)-1)] = view.PackageLine
	ll = append(ll, blankLine)
	ll = append(ll, suite.String())
	llMask[uint(len(ll)-1)] = view.SuiteLine
	suiteResults := p.OfSuite(suite)
	for i, out := range suiteResults.InitOut {
		if i == 0 {
			ll = append(ll, indent+"init-log:")
		}
		ll = append(ll, indent+indent+strings.TrimSpace(out))
	}
	if len(suiteResults.InitOut) > 0 {
		ll = append(ll, blankLine)
	}
	suite.ForTest(func(t *model.Test) {
		ll, llMask = reportSubResult(
			suiteResults.OfTest(t), indent, ll, llMask)
	})
	for i, out := range suiteResults.FinalizeOut {
		if i == 0 {
			ll = append(ll, blankLine, indent+"finalize-log:")
		}
		ll = append(ll, indent+indent+strings.TrimSpace(out))
	}
	return ll, llMask
}

func reportSuitesFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask, int, int) {
	var n, f int
	p.ForSuite(func(ts *model.TestSuite) {
		var _n, _f int
		ll, llMask, _n, _f = foldedSuiteLine(
			ts, p.OfSuite(ts), ll, llMask)
		n += _n
		f += _f
	})
	return ll, llMask, n, f
}

func reportGoTestsFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask, int, int) {
	n, f, d := 0, 0, time.Duration(0)
	p.ForTest(func(t *model.Test) {
		r := p.OfTest(t)
		n += r.Len()
		f += r.LenFailed()
		d += r.End.Sub(r.Start)
	})
	content := fmt.Sprintf("go-tests%s%d/%d %s",
		lines.LineFiller, n, f, d.Round(1*time.Millisecond))
	ll = append(ll, content)
	idx := uint(len(ll) - 1)
	llMask[idx] = view.GoTestsFoldedLine
	if f > 0 {
		llMask[idx] |= view.Failed
	}
	return ll, llMask, n, f
}

func reportGoTestsSubTestsFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	n, f, d, without, withSubs := goSplitTests(p)
	content := fmt.Sprintf("go-tests%s%d/%d %s",
		lines.LineFiller, n, f, d.Round(1*time.Millisecond))
	ll = append(ll, content)
	idx := uint(len(ll) - 1)
	llMask[idx] = view.GoTestsLine
	if f > 0 {
		llMask[idx] |= view.Failed
	}
	for _, t := range without {
		ll, llMask = reportResult(p.OfTest(t), indent, ll, llMask)
	}
	ll = append(ll, blankLine)
	for _, t := range withSubs {
		ll = append(ll, withFoldInfo(indent+t.String(), p.OfTest(t)))
		llMask[uint(len(ll)-1)] = view.GoSuiteFoldedLine
	}
	return ll, llMask
}
