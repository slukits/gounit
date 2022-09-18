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
	"github.com/slukits/lines"
)

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
	subInfo := map[uint]suiteInfo{}
	for _, t := range withSubs {
		tr := p.OfTest(t)
		ll = append(ll, "", t.Name())
		mask[uint(len(ll)-1)] = view.GoSuiteLine
		subInfo[uint(len(ll)-1)] = suiteInfo{
			ttN: tr.Len(), ffN: tr.LenFailed(),
			dr: time.Duration(tr.End.Sub(tr.Start))}
		ss := []*model.SubResult{}
		tr.For(func(sr *model.SubResult) {
			ss = append(ss, sr)
		})
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Name < ss[j].Name
		})
		for _, s := range ss {
			ll = append(ll, "    "+s.Name)
			mask[uint(len(ll)-1)] = view.SuiteTestLine
		}
	}
	return []interface{}{&goReport{
		report: report{
			flags: view.RpClearing,
			ll:    ll,
			mask:  mask,
		},
		suitesInfo: subInfo,
	}}
}

// goReport reports tests of a package only containing go tests.
type goReport struct {
	report
	suitesInfo map[uint]suiteInfo
	typ        reportType
	folded     *goReport
}

func (r *goReport) Type() reportType { return r.typ }

func (r *goReport) setType(t reportType) { r.typ = t }

func (r *goReport) Folded() reporter {
	if r.folded == nil {
		_r := &goReport{
			report: report{
				lst:   r.lst,
				flags: view.RpClearing,
				mask:  map[uint]view.LineMask{},
			},
			typ: rprGoOnlyFolded,
		}
		for idx, content := range r.ll {
			ux := uint(idx)
			if idx > 0 && content == "" {
				if r.LineMask(ux-1)&view.SuiteTestLine > 0 {
					continue
				}
			}
			if r.LineMask(ux)&view.SuiteTestLine > 0 {
				continue
			}
			if r.LineMask(ux)&view.GoSuiteLine > 0 {
				info := r.suitesInfo[ux]
				content = fmt.Sprintf("%s%s%d/%d %s",
					content, lines.LineFiller, info.ttN, info.ffN,
					info.dr.Round(1*time.Millisecond))
			}
			_r.ll = append(_r.ll, content)
			_r.mask[uint(len(_r.ll)-1)] = r.mask[ux]
		}
		r.folded = _r
	}
	return r.folded
}
