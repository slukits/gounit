// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"strings"
	"sync"
	"time"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
)

type modelState struct {
	*sync.Mutex
	current     []interface{}
	viewUpdater func(...interface{})
	pp          pkgs
	latest      string
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

func (s *modelState) update(pp pkgs, latest string) {
	s.Lock()
	defer s.Unlock()
	s.pp = pp
	s.latest = latest
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
	lm := s.current[0].(*report).LineMask(uint(idx))
	switch {
	case lm&view.PackageLine > 0:
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
		s.viewUpdater(s.current...)
		return
	}
	status := reportStatus(s.pp)
	ll, llMask, p := rprLines{}, linesMask{}, s.pp[s.latest]
	switch p.LenSuites() {
	case 0:
		ll, llMask = reportGoOnlyPkg(t, p, ll, llMask)
	default:
		ll, llMask = s.reportMixedPkg(t, idx, p, ll, llMask)
	}
	s.current = []interface{}{&report{
		flags:   view.RpClearing,
		ll:      ll,
		llMasks: llMask,
		lst:     s.lineListener,
	}, status}
	s.viewUpdater(s.current...)
}

func (s *modelState) reportMixedPkg(
	t reportType, idx int, p *pkg, ll rprLines, llMask linesMask,
) (rprLines, linesMask) {
	switch t {
	case rprGoTests:
		return reportMixedGoTests(p, ll, llMask)
	case rprSuite:
		var suite *model.TestSuite
		ln := s.current[0].(*report).ll[idx]
		p.ForSuite(func(ts *model.TestSuite) {
			if suite != nil {
				return
			}
			if strings.HasPrefix(ln, ts.Name()) {
				suite = ts
			}
		})
		return reportMixedSuite(suite, p, ll, llMask)
	case rprDefault:
		return reportMixedFolded(p, ll, llMask)
	}
	return ll, llMask
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
		for i := 0; i < n; i++ {
			p := <-rslt
			pp[p.ID()] = p
		}
		if len(pp) == 0 || pp[latest] == nil {
			return
		}
		mdl.update(pp, latest)
		mdl.report(rprDefault)
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
