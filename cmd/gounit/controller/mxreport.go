// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"sort"
	"time"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
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
	ll, llMask = reportFailedPkgSuitesHeader(p, ll, llMask)
	return ll, llMask
}

func reportMixedGoSuite(
	goSuite *model.Test, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
	ll = append(ll, blankLine)
	n, f, d, _, _ := goSplitTests(p)
	ll, llMask = reportGoTestsLine(
		n, f, d, view.GoTestsLine, ll, llMask)
	ll, llMask = reportFailedGoTests(p, ll, llMask)
	tr := p.OfTest(goSuite)
	ll, llMask = reportGoSuiteLine(
		tr, view.GoSuiteLine, indent, ll, llMask)
	tr.ForOrdered(func(sr *model.SubResult) {
		ll, llMask = reportSubTestLine(sr, indent+indent, ll, llMask)
	})
	return reportFailedSuitesBut(nil, p, ll, llMask)
}

func reportFailedGoTests(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	p.ForTest(func(t *model.Test) {
		if p.OfTest(t).Passed || p.OfTest(t).HasSubs() {
			return
		}
		ll, llMask = reportTestLine(p.OfTest(t), indent, ll, llMask)
	})

	return ll, llMask
}

func reportFailedSuitesBut(
	suite *model.TestSuite, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	var ss []*model.TestSuite
	p.ForSuite(func(ts *model.TestSuite) {
		if p.OfSuite(ts).Passed || (suite != nil && suite.Name() == ts.Name()) {
			return
		}
		ss = append(ss, ts)
	})
	if len(ss) == 0 {
		return ll, llMask
	}
	sort.Slice(ss, func(i, j int) bool {
		return ss[i].Name() < ss[j].Name()
	})

	for _, s := range ss {
		ll, llMask = reportSuiteLine(
			p, s, view.SuiteFoldedLine, ll, llMask)
	}
	return ll, llMask
}

func reportMixedSuite(
	suite *model.TestSuite, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
	ll = append(ll, blankLine)
	ll, llMask = reportFailedGoTestsHeader(p, ll, llMask)
	ll, llMask = reportFailedSuitesBut(suite, p, ll, llMask)
	ll, llMask = reportSuite(p, suite, ll, llMask)
	return ll, llMask
}

func reportFailedGoTestsHeader(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	n, f, d := 0, 0, time.Duration(0)
	p.ForTest(func(t *model.Test) {
		r := p.OfTest(t)
		n += r.Len()
		f += r.LenFailed()
		d += r.End.Sub(r.Start)
	})
	if f == 0 {
		return ll, llMask
	}

	reportGoTestsLine(
		n, f, d, view.GoTestsFoldedLine, ll, llMask)
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
