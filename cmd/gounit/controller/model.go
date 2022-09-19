// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"sync"

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
	s.report(s.reportTransition(idx))
}

// reportTransition calculates in dependency of the selected line and
// the current report-type the report-type for the user-request
// response.
func (s *modelState) reportTransition(idx int) reportType {
	s.Lock()
	defer s.Unlock()
	rpr := s.current[0].(*report)
	switch {
	case rpr.LineMask(uint(idx))&view.PackageLine > 0:
	case rpr.LineMask(uint(idx))&view.GoSuiteLine > 0:
		switch rpr.Type() {
		case rprGoOnly:
			return rprGoOnlyFolded
		case rprGoOnlyFolded:
			return rprGoOnly
		}
	}
	return rprDefault
}

func (s *modelState) report(t reportType) {
	s.Lock()
	defer s.Unlock()
	if t == rprCurrent {
		s.viewUpdater(s.current...)
		return
	}
	status := reportStatus(s.pp)
	ll, llMask, p := rprLines{}, linesMask{}, s.pp[s.latest]
	var rt reportType
	switch t {
	case rprDefault:
		switch p.tp.LenSuites() {
		case 0:
			ll, llMask = reportGoTestsOnly(p, ll, llMask)
			rt = rprGoOnly
		default:
		}
	case rprGoOnly:
		ll, llMask = reportGoTestsOnly(p, ll, llMask)
		rt = rprGoOnly
	case rprGoOnlyFolded:
		ll, llMask = reportGoTestsOnlyFolded(p, ll, llMask)
		rt = rprGoOnlyFolded
	}
	s.current = []interface{}{&report{
		typ:     rt,
		flags:   view.RpClearing,
		ll:      ll,
		llMasks: llMask,
		lst:     s.lineListener,
	}, status}
	s.viewUpdater(s.current...)
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
		mdl.update(pp, latest)
		mdl.report(rprDefault)
	}
}

func run(p *pkg, rslt chan *pkg) {
	rr, err := p.tp.Run()
	p.runResult = &runResult{Results: rr, err: err}
	rslt <- p
}

type runResult struct {
	err error
	*model.Results
}

type pkg struct {
	*runResult
	tp *model.TestingPackage
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
	// rprGoOnly reports one specific package
	rprGoOnly
	// rprGoOnlyFolded reports one specific package with folded
	// tests
	rprGoOnlyFolded
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
