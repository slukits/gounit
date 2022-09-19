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

const blankLine = ""

func reportGoTestsOnly(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	n, withoutSubs, withSubs := goSplitTests(p)
	ll, llMask = goWithoutSubs(p, ll, llMask, n, withoutSubs)

	for _, t := range withSubs {
		tr := p.OfTest(t)
		ll = append(ll, blankLine, t.Name())
		llMask[uint(len(ll)-1)] = view.GoSuiteLine
		tr.ForOrdered(func(sr *model.SubResult) {
			ll = append(ll, "    "+sr.Name)
			llMask[uint(len(ll)-1)] = view.SuiteTestLine
		})
	}
	return ll, llMask
}

func reportGoTestsOnlyFolded(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {

	n, withoutSubs, withSubs := goSplitTests(p)
	ll, llMask = goWithoutSubs(p, ll, llMask, n, withoutSubs)

	ll = append(ll, blankLine)
	for _, t := range withSubs {
		ll = append(ll, withFoldInfo(t.Name(), p.OfTest(t)))
		llMask[uint(len(ll)-1)] = view.GoSuiteLine
	}
	return ll, llMask
}

// goSplitTests splits go test into tests without and with sub-tests.
func goSplitTests(p *pkg) (n int, without, with []*model.Test) {
	p.tp.ForTest(func(t *model.Test) {
		if p.OfTest(t).Len() == 1 {
			without = append(without, t)
			n++
			return
		}
		n += p.OfTest(t).Len()
		with = append(with, t)
	})
	sort.Slice(without, func(i, j int) bool {
		return without[i].Name() < without[j].Name()
	})
	sort.Slice(with, func(i, j int) bool {
		return with[i].Name() < with[j].Name()
	})
	return n, without, with
}

// goWithoutSubs reports the package-line and the go-tests without
// sub-tests.
func goWithoutSubs(
	p *pkg, ll rprLines, llMask linesMask, n int, without []*model.Test,
) (rprLines, linesMask) {
	ll = append(ll, fmt.Sprintf("%s: %d/0 %v", p.tp.ID(), n,
		p.Duration.Round(1*time.Millisecond)))
	llMask[uint(len(ll)-1)] = view.PackageLine
	ll = append(ll, blankLine)
	for _, t := range without {
		ll = append(ll, t.Name())
		llMask[uint(len(ll)-1)] = view.TestLine
	}
	return ll, llMask
}
