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

// indent of a reported test-name.
const indent = "  "

func reportMixedFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
	ll = append(ll, blankLine)
	if p.LenTests() > 0 {
		ll, llMask = reportGoTestsFolded(p, ll, llMask)
	}
	ll, llMask = reportSuitesFolded(p, ll, llMask)
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
	ll, llMask = reportGoTestWithSubsFolded(p, ll, llMask)
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
		ll, llMask = reportSubTestLine(sr, indent+indent, ll, llMask)
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
	ll, llMask = reportSuite(p, suite, ll, llMask)
	return ll, llMask
}

func reportSuitesFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	p.ForSuite(func(ts *model.TestSuite) {
		ll, llMask = reportSuiteLine(
			p, ts, view.SuiteFoldedLine, ll, llMask)
	})
	return ll, llMask
}

func reportGoTestsFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	n, f, d := 0, 0, time.Duration(0)
	p.ForTest(func(t *model.Test) {
		r := p.OfTest(t)
		n += r.Len()
		f += r.LenFailed()
		d += r.End.Sub(r.Start)
	})
	ll, llMask = reportGoTestsLine(
		n, f, d, view.GoTestsFoldedLine, ll, llMask)
	return ll, llMask
}

func reportGoTestWithSubsFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	n, f, d, without, withSubs := goSplitTests(p)
	ll, llMask = reportGoTestsLine(n, f, d, view.GoTestsLine, ll, llMask)
	for _, t := range without {
		ll, llMask = reportTestLine(p.OfTest(t), indent, ll, llMask)
	}
	ll = append(ll, blankLine)
	for _, t := range withSubs {
		ll = append(ll, withFoldInfo(indent+t.String(), p.OfTest(t)))
		llMask[uint(len(ll)-1)] = view.GoSuiteFoldedLine
	}
	return ll, llMask
}
