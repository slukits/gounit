// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"sort"
	"strings"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
)

const blankLine = ""

func reportGoOnly(
	st *state, p *pkg, rt reportType, idx int,
	ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	switch rt {
	case rprGoSuite:
		gs := findGoSuite(st, p, idx)
		if gs == nil {
			return reportPackages(st.pp)
		}
		st.lastSuite = "go-tests:" + gs.Name()
		ll, llMask = reportGoOnlySuite(p, gs, ll, llMask)
	case rprGoSuiteFolded:
		st.lastSuite = ""
		fallthrough
	default:
		if strings.HasPrefix(st.lastSuite, "go-tests:") {
			return reportLockedGoSuite(st, p, ll, llMask)
		}
		var goSuite string
		ll, llMask, goSuite = reportGoOnlyPkg(p, rt, ll, llMask)
		if goSuite != "" {
			st.lastSuite = "go-tests:" + goSuite
		}
	}
	return ll, llMask
}

func reportGoOnlyPkg(
	p *pkg, rt reportType, ll rprLines, llMask linesMask,
) (rprLines, linesMask, string) {

	n, f, withoutSubs, withSubs := goSplitTests(p)

	folded := func() (rprLines, linesMask, string) {
		ll, llMask = goWithoutSubs(p, ll, llMask, n, f, withoutSubs)
		ll = append(ll, blankLine)
		for _, t := range withSubs {
			ll, llMask = reportGoSuiteLine(
				p.OfTest(t), view.GoSuiteFoldedLine, "", ll, llMask)
		}
		return ll, llMask, ""
	}

	if withoutSubs.haveFailed(p) {
		return folded()
	}

	var fld *model.Test
	for _, t := range withSubs {
		if p.OfTest(t).Passed {
			continue
		}
		fld = t
	}
	if fld != nil && !userRequestsParticular(rt) {
		ll, llMask = reportGoOnlySuite(p, fld, ll, llMask)
		return ll, llMask, fld.Name()
	}

	return folded()
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

func reportLockedGoSuite(
	st *state, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	if p.LenTests() == 0 {
		return reportMixedFolded(p, ll, llMask)
	}
	if st.lastSuite == "go-tests" {
		return reportMixedGoTests(p, ll, llMask)
	}
	goSuiteName := strings.Split(st.lastSuite, ":")[1]
	var goSuite *model.Test
	p.ForTest(func(t *model.Test) {
		if goSuite != nil {
			return
		}
		if t.Name() != goSuiteName {
			return
		}
		goSuite = t
	})
	if goSuite != nil {
		return reportMixedGoSuite(goSuite, p, ll, llMask)
	}
	return reportMixedFolded(p, ll, llMask)
}

func findGoSuite(st *state, p *pkg, idx int) *model.Test {
	var goSuite *model.Test
	ln, ok := findReportLine(st.view[0].(*report), idx,
		view.GoSuiteFoldedLine|view.GoSuiteLine)
	if !ok {
		return nil
	}
	p.ForTest(func(t *model.Test) {
		if goSuite != nil {
			return
		}
		if ln == t.String() {
			goSuite = t
		}
	})
	return goSuite
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
	n, f int, without, with goTests,
) {
	p.ForTest(func(t *model.Test) {
		r := p.OfTest(t)
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
	return n, f, without, with
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
