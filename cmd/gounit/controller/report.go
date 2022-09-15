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
	tm time.Duration
}

type pkg struct {
	*runResult
	tp *model.TestingPackage
}

type pkgs map[string]*pkg

func watch(
	watched <-chan *model.PackagesDiff,
	vwUpd func(...interface{}),
) {
	pp := pkgs{}
	for diff := range watched {
		if diff == nil {
			return
		}
		rslt, n := make(chan *pkg), 0
		var latest string
		diff.For(func(tp *model.TestingPackage) (stop bool) {
			n++
			go run(&pkg{tp: tp}, rslt)
			latest = tp.ID()
			return
		})
		for i := 0; i < n; i++ {
			p := <-rslt
			pp[p.tp.ID()] = p
		}
		if len(pp) == 0 || pp[latest] == nil {
			return
		}
		report(pp, latest, vwUpd)
	}
}

func run(p *pkg, rslt chan *pkg) {
	rr, err := p.tp.Run()
	p.runResult = &runResult{Results: rr, err: err}
	rslt <- p
}

func report(pp pkgs, latest string, vwUpd func(...interface{})) {
	if pp[latest].tp.LenSuites() == 0 {
		reportGoTestsOnly(pp, latest, vwUpd)
		return
	}
}

func reportGoTestsOnly(
	pp pkgs, latest string, vwUpd func(...interface{}),
) {
	var singles, withSubs []*model.Test
	p, n, ll := pp[latest], 0, []string{}
	p.tp.ForTest(func(t *model.Test) {
		if p.OfTest(t).Len() == 1 {
			singles = append(singles, t)
			n++
			return
		}
		n += pp[latest].OfTest(t).Len()
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
	vwUpd(&reporter{flags: view.RpClearing, ll: ll})
}
