// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"sort"
	"time"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
)

const blankLine = ""

func reportGoOnlyPkg(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	n, f, _, withoutSubs, withSubs := goSplitTests(p)
	ll, llMask = goWithoutSubs(p, ll, llMask, n, f, withoutSubs)

	ll = append(ll, blankLine)
	for _, t := range withSubs {
		ll, llMask = reportGoSuiteLine(
			p.OfTest(t), view.GoSuiteFoldedLine, "", ll, llMask)
	}
	return ll, llMask
}

func reportGoOnlySuite(
	p *pkg, s *model.Test, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
	ll = append(ll, blankLine)
	tr := p.OfTest(s)
	ll, llMask = reportGoSuiteLine(
		tr, view.GoSuiteLine, "", ll, llMask)
	tr.ForOrdered(func(sr *model.SubResult) {
		ll, llMask = reportSubTestLine(p, sr, indent+indent, ll, llMask)
	})
	return ll, llMask
}

type goTests []*model.Test

func (tt goTests) haveFailed(p *pkg) bool {
	for _, t := range tt {
		if p.OfTest(t).Passed {
			continue
		}
		return true
	}
	return false
}

// goSplitTests splits go test into tests without and with sub-tests.
func goSplitTests(p *pkg) (
	n, f int, d time.Duration, without, with goTests,
) {
	p.ForTest(func(t *model.Test) {
		r := p.OfTest(t)
		if r == nil {
			return
		}
		d += r.End.Sub(r.Start)
		if r.HasSubs() {
			n += p.OfTest(t).Len()
			f += p.OfTest(t).LenFailed()
			with = append(with, t)
			return
		}
		without = append(without, t)
		n++
		if !r.Passed {
			f++
		}
	})
	sort.Slice(without, func(i, j int) bool {
		return without[i].Name() < without[j].Name()
	})
	sort.Slice(with, func(i, j int) bool {
		return with[i].Name() < with[j].Name()
	})
	return n, f, d, without, with
}

// goWithoutSubs reports the package-line and the go-tests without
// sub-tests.
func goWithoutSubs(
	p *pkg, ll rprLines, llMask linesMask, n, f int, without []*model.Test,
) (rprLines, linesMask) {
	ll, llMask = reportPackageLine(p, view.PackageLine, ll, llMask)
	ll = append(ll, blankLine)
	for _, t := range without {
		ll, llMask = reportTestLine(p, p.OfTest(t), "", ll, llMask)
	}
	return ll, llMask
}
