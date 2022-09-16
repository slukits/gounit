// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"sort"

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

func lastReport(pp pkgs, latest string) []interface{} {
	if pp[latest].tp.LenSuites() == 0 {
		return reportGoTestsOnly(pp[latest])
	}
	return nil
}

func reportGoTestsOnly(p *pkg) []interface{} {
	var singles, withSubs []*model.Test
	n, ll := 0, []string{}
	p.tp.ForTest(func(t *model.Test) {
		if p.OfTest(t).Len() == 1 {
			singles = append(singles, t)
			n++
			return
		}
		n += p.OfTest(t).Len()
		withSubs = append(withSubs, t)
	})
	ll = append(ll, fmt.Sprintf("%s: %d/0 %v", p.tp.ID(), n, p.Duration), "")
	sort.Slice(singles, func(i, j int) bool {
		return singles[i].Name() < singles[j].Name()
	})
	for _, t := range singles {
		ll = append(ll, t.Name())
	}
	sort.Slice(withSubs, func(i, j int) bool {
		return withSubs[i].Name() < withSubs[j].Name()
	})
	for _, t := range withSubs {
		tr := p.OfTest(t)
		ll = append(ll, "", t.Name())
		ss := []*model.SubResult{}
		tr.For(func(sr *model.SubResult) {
			ss = append(ss, sr)
		})
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Name < ss[j].Name
		})
		for _, s := range ss {
			ll = append(ll, "    "+s.Name)
		}
	}
	return []interface{}{&reporter{flags: view.RpClearing, ll: ll}}
}
