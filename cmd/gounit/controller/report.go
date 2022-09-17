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

type runResult struct {
	err error
	*model.Results
}

type pkg struct {
	*runResult
	tp *model.TestingPackage
}

type pkgs map[string]*pkg

func reportTestingPackage(p *pkg) []interface{} {
	if p.tp.LenSuites() == 0 {
		return reportGoTestsOnly(p)
	}
	return nil
}

func reportGoTestsOnly(p *pkg) []interface{} {
	var singles, withSubs []*model.Test
	n, ll, mask := 0, []string{}, map[uint]view.LineMask{}
	p.tp.ForTest(func(t *model.Test) {
		if p.OfTest(t).Len() == 1 {
			singles = append(singles, t)
			n++
			return
		}
		n += p.OfTest(t).Len()
		withSubs = append(withSubs, t)
	})
	ll = append(ll, fmt.Sprintf("%s: %d/0 %v", p.tp.ID(), n,
		p.Duration.Round(1*time.Millisecond)), "")
	mask[0] = view.PackageLine
	sort.Slice(singles, func(i, j int) bool {
		return singles[i].Name() < singles[j].Name()
	})
	for _, t := range singles {
		ll = append(ll, t.Name())
		mask[uint(len(ll)-1)] = view.TestLine
	}
	sort.Slice(withSubs, func(i, j int) bool {
		return withSubs[i].Name() < withSubs[j].Name()
	})
	for _, t := range withSubs {
		tr := p.OfTest(t)
		ll = append(ll, "", t.Name())
		mask[uint(len(ll)-1)] = view.SuiteLine
		ss := []*model.SubResult{}
		tr.For(func(sr *model.SubResult) {
			ss = append(ss, sr)
		})
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Name < ss[j].Name
		})
		for _, s := range ss {
			ll = append(ll, "    "+s.Name)
			mask[uint(len(ll)-1)] = view.TestLine
		}
	}
	return []interface{}{&reporter{
		flags: view.RpClearing, ll: ll, mask: mask}}
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
			rr := p.Results.OfSuite(ts)
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
