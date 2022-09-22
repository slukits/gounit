// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

type state struct {
	view        []interface{}
	pp          pkgs
	ee          []*pkg
	latestPkg   string
	latestSuite string
}

type modelState struct {
	*sync.Mutex
	*state
	viewUpdater func(...interface{})
}

func (s *modelState) replaceViewUpdater(f func(...interface{})) {
	s.Lock()
	defer s.Unlock()
	s.viewUpdater = f
}

func (s *modelState) clone() pkgs {
	s.Lock()
	defer s.Unlock()
	pp := pkgs{}
	for k, v := range s.pp {
		pp[k] = v
	}
	return pp
}

type reporter interface {
	setListener(func(int))
	Type() reportType
	setType(i reportType)
	LineMask(uint) view.LineMask
	Folded() reporter
}

func (s *modelState) update(st *state) {
	s.Lock()
	s.state = st
	if len(s.ee) > 0 {
		status := reportStatus(s.pp)
		s.view = []interface{}{
			reportFailed(s.state, s.lineListener), status}
		s.viewUpdater(s.view...)
	}
	s.Unlock()
	if len(s.ee) > 0 {
		return
	}
	s.report(rprDefault)
}

func (s *modelState) lineListener(idx int) {
	s._report(s.reportTransition(idx), idx)
}

// reportTransition calculates in dependency of the current report's
// selected line and type the new report-type for the user-request
// response.
func (s *modelState) reportTransition(idx int) reportType {
	s.Lock()
	defer s.Unlock()
	lm := s.view[0].(*report).LineMask(uint(idx))
	switch {
	case lm&view.PackageLine > 0:
		return rprPackages
	case lm&view.GoSuiteLine > 0:
		return rprGoSuiteFolded
	case lm&view.GoSuiteFoldedLine > 0:
		return rprGoSuite
	case lm&view.GoTestsFoldedLine > 0:
		return rprGoTests
	case lm&view.GoTestsLine > 0:
		return rprDefault
	case lm&view.SuiteFoldedLine > 0:
		return rprSuite
	case lm&view.SuiteLine > 0:
		return rprDefault
	case lm&view.PackageFoldedLine > 0:
		return rprPackage
	}
	return rprDefault
}

// report reports a report of given type.
func (s *modelState) report(t reportType) {
	s._report(t, -1) // panics if we have a bug
}

// _report creates report of given type.  The index is needed for the
// use case that a folded suite was selected to determine which one.
func (s *modelState) _report(t reportType, idx int) {
	s.Lock()
	defer s.Unlock()
	if t == rprCurrent {
		s.viewUpdater(s.view...)
		return
	}
	status := reportStatus(s.pp)
	if t == rprPackages {
		s.view = []interface{}{
			reportPackages(s.pp, s.lineListener), status}
		s.viewUpdater(s.view...)
		return
	}
	if t == rprPackage {
		pNm := strings.Split(s.view[0].(*report).ll[idx],
			lines.LineFiller)[0]
		for _, p := range s.pp {
			if p.ID() != pNm {
				continue
			}
			s.latestPkg = p.ID()
			break
		}
	}
	ll, llMask, p := rprLines{}, linesMask{}, s.pp[s.latestPkg]
	if len(s.ee) > 0 {
		ll, llMask = s.reportFailedPkgsBut(p, ll, llMask)
	}
	if p.HasErr() {
		ll, llMask = reportFailedPkg(p, ll, llMask)
		s.view = []interface{}{&report{
			flags:   view.RpClearing,
			ll:      ll,
			llMasks: llMask,
			lst:     s.lineListener,
		}, status}
		s.viewUpdater(s.view...)
		return
	}
	switch p.LenSuites() {
	case 0:
		switch t {
		case rprGoSuite:
			ll, llMask = reportGoOnlySuite(
				p, s.findGoSuite(p, idx), ll, llMask)
		default:
			ll, llMask = reportGoOnlyPkg(p, ll, llMask)
		}
	default:
		ll, llMask = s.reportMixedPkg(t, idx, p, ll, llMask)
	}
	s.view = []interface{}{&report{
		flags:   view.RpClearing,
		ll:      ll,
		llMasks: llMask,
		lst:     s.lineListener,
	}, status}
	s.viewUpdater(s.view...)
}

func (s *modelState) reportFailedPkgsBut(
	p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	sort.Slice(s.ee, func(i, j int) bool {
		return s.ee[i].ID() < s.ee[j].ID()
	})
	for _, ep := range s.ee {
		if ep.ID() == p.ID() {
			continue
		}
		ll, llMask = reportPackageLine(
			ep, view.PackageFoldedLine, ll, llMask)
	}
	return ll, llMask
}

func (s *modelState) reportMixedPkg(
	t reportType, idx int, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	switch t {
	case rprGoTests, rprGoSuiteFolded:
		return reportMixedGoTests(p, ll, llMask)
	case rprSuite:
		var suite *model.TestSuite
		ln := strings.Split(s.view[0].(*report).ll[idx],
			lines.LineFiller)[0]
		p.ForSuite(func(ts *model.TestSuite) {
			if suite != nil {
				return
			}
			if ln == ts.String() {
				suite = ts
			}
		})
		return reportMixedSuite(suite, p, ll, llMask)
	case rprGoSuite:
		return reportMixedGoSuite(
			s.findGoSuite(p, idx), p, ll, llMask)
	case rprDefault, rprPackage:
		return reportMixedFolded(p, ll, llMask)
	}
	return ll, llMask
}

func (s *modelState) findGoSuite(p *pkg, idx int) *model.Test {
	var goSuite *model.Test
	ln := strings.TrimSpace(strings.Split(
		s.view[0].(*report).ll[idx], lines.LineFiller)[0])
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

func watch(
	watched <-chan *model.PackagesDiff,
	mdl *modelState,
) {
	for diff := range watched {
		if diff == nil {
			return
		}
		pp, latest := mdl.clone(), ""
		rslt, n := make(chan *pkg), 0
		diff.For(func(tp *model.TestingPackage) (stop bool) {
			n++
			go run(&pkg{TestingPackage: tp}, rslt)
			latest = tp.ID()
			return
		})
		ee := []*pkg{}
		for i := 0; i < n; i++ {
			p := <-rslt
			pp[p.ID()] = p
			if !p.HasErr() && p.Passed() {
				continue
			}
			ee = append(ee, p)
		}
		if len(pp) == 0 || pp[latest] == nil {
			return
		}
		mdl.update(&state{pp: pp, ee: ee, latestPkg: latest})
	}
}

func run(p *pkg, rslt chan *pkg) {
	rr, err := p.Run()
	p.runResult = &runResult{Results: rr, err: err}
	rslt <- p
}

type runResult struct {
	err error
	*model.Results
}

type pkg struct {
	*runResult
	*model.TestingPackage
}

// info counts a package's tests, failed tests and the provides the
// (actual) duration of the package's test run.
func (p *pkg) info() (n, f int, d time.Duration) {
	p.ForTest(func(t *model.Test) {
		r := p.OfTest(t)
		n += r.Len()
		f += r.LenFailed()
	})
	p.ForSuite(func(st *model.TestSuite) {
		r := p.OfSuite(st)
		n += r.Len()
		f += r.LenFailed()
	})
	return n, f, p.Duration
}

type pkgs map[string]*pkg

// reportType values type model-state reports which is leveraged
// for transitioning between different views.  E.g. a click on a package
// name should show all packages if the type is not rprPackages; iff the
// type is rprPackages than the clicked package should be reported.
type reportType int

const (
	// rprDefault reports the initial view of a model-state
	rprDefault reportType = iota
	// rprCurrent re-reports the currently reported report; e.g. the
	// user "closes" the help screen.
	rprCurrent
	// rprGoSuite reports a package having only go-tests reporting all
	// tests and sub-tests.
	rprGoSuite
	// rprGoTests report go tests unfolded (with folded sub-tests) of a
	// mixed package.
	rprGoTests
	// rprSuite report a particular suite of a package.
	rprSuite
	// rprGoSuiteFolded reports a package having only go-tests with folded
	// sub-tests.
	rprGoSuiteFolded
	// rprMixedFolded reports a package consisting of test-suites and
	// optionally go-tests with all suites and go-tests folded.
	rprMixedFolded
	// rprPackages reports all packages of the watched directory
	rprPackages
	// rprPackage reports a specific package
	rprPackage
	// rprPackage reports a single package with all suites folded
	rprPackageFolded
	// rprPackageFocusedGo reports a single package's go tests
	rprPackageFocusedGo
	// rprPackageFocusedGoFolded reports a single package's go tests
	// with folded sub-tests
	rprPackageFocusedGoFolded
)

const (
	WatcherErr = "gounit: watcher: %s: %v"
)

// A Watcher implementation provides the needed information about a
// watched source directory to the controller.
type Watcher interface {

	// ModuleName is the module name of the descendant watched source
	// directory.
	ModuleName() string

	// ModuleDir returns the absolute path of the module directory of
	// the descendant watched source directory.
	ModuleDir() string

	// SourcesDir returns the absolute path of the watched source
	// directory.
	SourcesDir() string

	// Watch is a function whose returned channel watches a go modules
	// packages sources whose tests runs are reported to a terminal ui.
	Watch() (<-chan *model.PackagesDiff, uint64, error)
}
