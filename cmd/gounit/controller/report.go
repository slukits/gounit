// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"time"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

// report is the simplest implementation of view.Reporter.
type report struct {
	flags view.RprtMask
	ll    []string
	mask  map[uint]view.LineMask
	lst   func(int)
}

// Clearing indicates if all lines not set by this reporter's For
// function should be cleared or not.
func (r *report) Flags() view.RprtMask { return r.flags }

// For expects the view's reporting component and a callback to which
// the updated lines can be provided to.
func (r *report) For(_ lines.Componenter, line func(uint, string)) {
	for idx, content := range r.ll {
		line(uint(idx), content)
	}
}

// Mask returns for given index special formatting directives.
func (r *report) LineMask(idx uint) view.LineMask {
	if r.mask == nil {
		return view.ZeroLineMod
	}
	return r.mask[idx]
}

// Listener returns the callback which is informed about user selections
// of lines by providing the index of the selected line.
func (r *report) Listener() func(int) { return r.lst }

// setListener is part of the reporter-implementation.
func (r *report) setListener(l func(int)) {
	r.lst = l
}

func reportTestingPackage(p *pkg) []interface{} {
	if p.tp.LenSuites() == 0 {
		return reportGoTestsOnly(p)
	}
	return nil
}

type suiteInfo struct {
	ttN, ffN int
	dr       time.Duration
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
			rr := p.OfSuite(ts)
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
