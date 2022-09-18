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
	lastReport  []interface{}
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

func (s *modelState) update(
	pp pkgs, latest string, lastReport []interface{},
) {
	lastReport[0].(reporter).setListener(s.lineListener)
	s.Lock()
	defer s.Unlock()
	s.pp = pp
	s.latest = latest
	s.lastReport = lastReport
}

func (s *modelState) lineListener(idx int) {
	s.Lock()
	defer s.Unlock()
	rpr, ok := s.lastReport[0].(reporter)
	if !ok {
		panic("first component of last report must be reporter")
	}
	lm := view.ZeroLineMod
	switch rpr.Type() {
	case rprGoOnly:
		lm = rpr.LineMask(uint(idx))
	case rprGoOnlyFolded:
		lm = rpr.Folded().LineMask(uint(idx))
	}
	switch {
	case lm&view.PackageLine > 0:
	case lm&view.GoSuiteLine > 0:
		if rpr.Type() == rprGoOnly {
			rpr.setType(rprGoOnlyFolded)
			s.viewUpdater(append([]interface{}{
				rpr.Folded()}, s.lastReport[1:]...)...)
			return
		}
		rpr.setType(rprGoOnly)
		s.viewUpdater(s.lastReport...)
		return
	}
}

func (s *modelState) report() {
	s.Lock()
	defer s.Unlock()
	s.viewUpdater(s.lastReport...)
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
		lastReport := reportTestingPackage(pp[latest])
		lastReport = append(lastReport, reportStatus(pp))
		mdl.update(pp, latest, lastReport)
		mdl.report()
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
	// rprGoOnly reports one specific package
	rprGoOnly reportType = iota
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
