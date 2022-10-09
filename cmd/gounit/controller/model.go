// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"sync"
	"time"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
)

// state represents the current state of a watched source directory from
// gounit's point of view, i.e. directories which are testing go
// packages and the results of their test runs.
type state struct {

	// pp are the testing packages of a watched directory.
	pp pkgs

	// ee are the packages of pp which have errors or failing tests.
	ee map[string]bool

	// view holds the current view needed if an other view like help or
	// about is shown and the use wants to go back to reported testing
	// packages.
	view []interface{}

	isOn onMask

	// stillTheSame indicates a user input referring to a stale state.
	stillTheSame bool
	latestPkg    string
	lastSuite    string
}

// ensureLatestPackage determines the latestPkg, i.e. the package to
// report.
func (s *state) ensureLatestPackage() *pkg {
	if s.latestPkg != "" && s.pp[s.latestPkg] != nil {
		return s.pp[s.latestPkg]
	}
	if s.latestPkg != "" {
		s.lastSuite = ""
	}
	if len(s.ee) > 0 {
		latest := ""
		for e := range s.ee {
			if latest == "" {
				latest = e
				continue
			}
			if s.pp[latest].ModTime.Before(s.pp[e].ModTime) {
				latest = e
			}
		}
		s.latestPkg = latest
		return s.pp[s.latestPkg]
	}
	p := s.pp.latest()
	s.latestPkg = p.ID()
	return p
}

// modelState manages the current state of a watched source directory's
// testing packages.
type modelState struct {

	// Mutex protects a change in state calculated by watch due to a
	// testing package update and it protects a change of the
	// view-property which may be a result of a state change or a user
	// input.
	*sync.Mutex

	// state represents the current state of testing packages of a
	// watched source directory
	*state

	// viewUpdater provides serialized access to the view.
	viewUpdater func(...interface{})

	// msgUpdater is a closure which calculates the update of the view's
	// message bar in case of an state change.
	msgUpdater func(string) string

	// isSuspended controls if model-state change updates the view.
	isSuspended bool
}

func msgUpdater(mdlName, srcDir string) func(string) string {
	return func(s string) string {
		if s == "" || s == srcDir {
			return ""
		}
		return fmt.Sprintf("%s: %s", mdlName, s)
	}
}

func (s *modelState) suspend() {
	s.Lock()
	defer s.Unlock()
	s.isSuspended = true
}

func (s *modelState) resume() {
	s.Lock()
	s.isSuspended = false
	s.Unlock()
	s.report(rprCurrent)
}

func (s *modelState) replaceViewUpdater(f func(...interface{})) {
	s.Lock()
	defer s.Unlock()
	s.viewUpdater = f
}

// clone returns a shallow copy of given model-state's state allowing to
// change the state or report a different aspect of it with a minimum of
// locking.
func (s *modelState) clone(switchStillTheSame bool) *state {
	s.Lock()
	defer s.Unlock()
	pp := pkgs{}
	for k, v := range s.pp {
		pp[k] = v
	}
	ee := map[string]bool{}
	for k, v := range s.ee {
		ee[k] = v
	}
	view := append([]interface{}{}, s.view...)
	if switchStillTheSame {
		s.stillTheSame = true
	}
	return &state{pp: pp, ee: ee, view: view, latestPkg: s.latestPkg,
		lastSuite: s.lastSuite, isOn: s.isOn}
}

type reporter interface {
	setListener(func(int))
	Type() reportType
	setType(i reportType)
	LineMask(uint) view.LineMask
	Folded() reporter
}

// updateState through a change in the watched source directory.
func (s *modelState) updateState(st *state) {
	if st.latestPkg != "" && len(st.ee) > 0 && !st.ee[st.latestPkg] {
		st.latestPkg = ""
	}
	status := newStatus(st.pp, st.isOn)
	r := newReport(st, rprDefault, -1)
	s.Lock()
	defer s.Unlock()
	s.state = st
	s.updateView(st, r, status)
}

// lineListener is called back from the view iff a user-input selected a
// line of the report component.
func (s *modelState) lineListener(idx int) {
	st := s.clone(true)
	rt := reportTransition(st, idx)
	update := func() {
		status := newStatus(st.pp, s.isOn)
		report := newReport(st, rt, idx)
		s.updateReport(st, report, status)
	}
	if rt == rprPackage && ensureRequestedPackageRun(update, st, idx) {
		return
	}
	update()
}

// ensureRequestedPackageRun makes sure that a user-selected package's
// tests have been run according to the set vet/race flags and returns
// true if a package's tests were rerun.  This function covers the use
// case if the user requests all packages folded, then for example turns
// vetting on and finally selects a package to report.
func ensureRequestedPackageRun(cb func(), st *state, idx int) bool {
	pID, ok := findReportLine(
		st.view[0].(*report), idx, view.PackageFoldedLine)
	if !ok {
		return false
	}
	pkg := st.pp[pID]
	if pkg.om == st.isOn {
		return false
	}
	go rerunTests(cb, pkg, st)
	return true
}

// reportTransition calculates in dependency of the current report's
// selected line and type the new report-type for the user-request
// response.
func reportTransition(st *state, idx int) reportType {
	lm := st.view[0].(*report).LineMask(uint(idx))
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
		return rprPackageFolded
	case lm&view.SuiteFoldedLine > 0:
		return rprSuite
	case lm&view.SuiteLine > 0:
		return rprMixedFolded
	case lm&view.PackageFoldedLine > 0:
		return rprPackage
	}
	return rprDefault
}

// report reports a report of given type.
func (s *modelState) report(t reportType) {
	if t == rprCurrent {
		s.Lock()
		defer s.Unlock()
		if s.isSuspended {
			return
		}
		s.viewUpdater(s.view...)
		return
	}
	st := s.clone(true)
	status := newStatus(st.pp, s.isOn)
	report := newReport(st, t, -1)
	s.updateReport(st, report, status)
}

func (s *modelState) setOnFlag(om onMask) {
	st := s.clone(true)
	st.isOn |= om

	if st.latestPkg != "" && om&(raceOn|vetOn) != 0 {
		go rerunTests(func() {
			stt := newStatus(st.pp, st.isOn)
			r := newReport(st, rprDefault, -1)
			s.updateReport(st, r, stt)
		}, st.pp[st.latestPkg], st)
		return
	}

	stt := newStatus(st.pp, st.isOn)
	var r *report
	if st.latestPkg != "" {
		r = newReport(st, rprDefault, -1)
	}
	if st.latestPkg == "" {
		r = newReport(st, rprPackages, -1)
	}
	s.updateReport(st, r, stt)
}

// updateReport updates the currently reported report iff during report
// calculation the model-state has not been updated through a source
// change.
func (s *modelState) updateReport(
	st *state, report *report, status *view.Statuser,
) {
	s.Lock()
	defer s.Unlock()
	if !s.stillTheSame {
		// state was updated meanwhile => discard user input
		return
	}
	s.state = st
	s.updateView(st, report, status)
}

// updateView is the final step to get a new report to the view.  It
// adds if needed an update of the message bar.
func (s *modelState) updateView(
	st *state, report *report, status *view.Statuser,
) {
	report.lst = s.lineListener
	report.flags = view.RpClearing
	st.view = []interface{}{report, status, s.msgUpdater(st.latestPkg)}
	if s.isSuspended {
		return
	}
	s.viewUpdater(st.view...)
}

func (s *modelState) removeOneFlag(om onMask) {
	s.Lock()
	defer s.Unlock()
	s.isOn &^= om
	if om&statsOn != 0 {
		for _, p := range s.pp {
			p.ResetSrcStats()
		}
	}
}

// rerunTests reruns the tests of given package using the given state's
// isOn property to calculate the (usually modified) run-arguments for
// the test-run.  Finally given model-states report is updated with the
// rerun package's report.
func rerunTests(cb func(), p *pkg, st *state) {
	rr, err := p.Run(translateToRunMask(st.isOn))
	p.runResult = &runResult{Results: rr, err: err, om: st.isOn}
	if p.HasErr() && !st.ee[p.ID()] {
		st.ee[p.ID()] = true
	}
	if !p.HasErr() && st.ee[p.ID()] {
		delete(st.ee, p.ID())
	}
	p.inf = nil
	cb()
}

func translateToRunMask(om onMask) model.RunMask {
	rm := model.RunMask(0)
	if om&vetOn != 0 {
		rm |= model.RunVet
	}
	if om&raceOn != 0 {
		rm |= model.RunRace
	}
	return rm
}

func watch(
	watched <-chan *model.PackagesDiff,
	mdl *modelState,
	afterUpdate chan bool,
) {
	for diff := range watched {
		if diff == nil {
			return
		}
		st := mdl.clone(false)
		rslt, n := make(chan *pkg), 0
		// TODO: since we don't care about the reported package order we
		// should be able to remove the sorting of them from the model.
		diff.For(func(tp *model.TestingPackage) (stop bool) {
			n++ // count expected results
			go run(&pkg{TestingPackage: tp}, st.isOn, rslt)
			return
		})
		if st.ee == nil {
			st.ee = map[string]bool{}
		}
		for i := 0; i < n; i++ {
			p := <-rslt
			st.pp[p.ID()] = p
			if !p.HasErr() && p.Passed() {
				if st.ee[p.ID()] {
					delete(st.ee, p.ID())
				}
				continue
			}
			st.ee[p.ID()] = true
		}
		diff.ForDel(func(tp *model.TestingPackage) (stop bool) {
			delete(st.pp, tp.ID())
			delete(st.ee, tp.ID())
			return
		})
		mdl.updateState(st)
		if afterUpdate != nil {
			afterUpdate <- true
		}
	}
}

func run(p *pkg, om onMask, rslt chan *pkg) {
	rr, err := p.Run(translateToRunMask(om))
	p.TrimTo(rr)
	p.runResult = &runResult{Results: rr, err: err, om: om}
	rslt <- p
}

type runResult struct {
	// err represents a possible error of a failed tests-run
	err error
	// Results are the parsed test-results of a tests-run
	*model.Results
	// om documents the set flags of tests-run
	om onMask
}

type info struct {
	n, f, s int
	d       time.Duration
}

type pkg struct {
	*runResult
	*model.TestingPackage
	inf *info
}

// info counts a package's tests, failed tests, the number of suites and
// provides the (actual) duration of the package's test run.
func (p *pkg) info() (n, f, s int, d time.Duration) {
	if p.HasErr() {
		return 0, 0, 0, 0
	}
	if p.inf == nil {
		goSuites := 0
		p.ForTest(func(t *model.Test) {
			r := p.OfTest(t)
			n += r.Len()
			f += r.LenFailed()
			if r.HasSubs() {
				goSuites++
			}
		})
		p.ForSuite(func(st *model.TestSuite) {
			r := p.OfSuite(st)
			if r == nil {
				return
			}
			n += r.Len()
			f += r.LenFailed()
		})
		p.inf = &info{n: n, f: f, d: p.Duration,
			s: p.LenSuites() + goSuites}
	}
	return p.inf.n, p.inf.f, p.inf.s, p.inf.d
}

func (p *pkg) HasFailedSuite() bool {
	failed := false
	p.ForSuite(func(ts *model.TestSuite) {
		if failed {
			return
		}
		if p.OfSuite(ts).Passed {
			return
		}
		failed = true
	})
	return failed
}

type pkgs map[string]*pkg

func (pp pkgs) latest() *pkg {
	var l *pkg
	for _, p := range pp {
		if l == nil {
			l = p
			continue
		}
		if l.ModTime.Before(p.ModTime) {
			l = p
		}
	}
	return l
}

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
