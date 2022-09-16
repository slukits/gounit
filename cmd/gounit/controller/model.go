// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"sync"

	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/gounit/cmd/gounit/view"
	"github.com/slukits/lines"
)

type modelState struct {
	*sync.Mutex
	lastReport  []interface{}
	viewUpdater func(...interface{})
	pp          pkgs
	latest      string
}

func (s *modelState) clone() *modelState {
	s.Lock()
	defer s.Unlock()
	ms := &modelState{
		Mutex:       &sync.Mutex{},
		viewUpdater: s.viewUpdater,
		pp:          pkgs{},
	}
	for k, v := range s.pp {
		ms.pp[k] = v
	}
	return ms
}

func (s *modelState) update(ms *modelState) {
	_ms := *s
	_ms.Lock()
	*s = *ms
	_ms.Unlock()
}

func (s *modelState) report() {
	if s.lastReport == nil {
		if s.pp[s.latest].tp.LenSuites() == 0 {
			s.lastReport = reportGoTestsOnly(s.pp, s.latest)
		}
	}
	s.viewUpdater(s.lastReport...)
}

func watch(
	watched <-chan *model.PackagesDiff,
	mdl *modelState,
) {
	for diff := range watched {
		m := mdl.clone()
		if diff == nil {
			return
		}
		rslt, n := make(chan *pkg), 0
		diff.For(func(tp *model.TestingPackage) (stop bool) {
			n++
			go run(&pkg{tp: tp}, rslt)
			m.latest = tp.ID()
			return
		})
		for i := 0; i < n; i++ {
			p := <-rslt
			m.pp[p.tp.ID()] = p
		}
		if len(m.pp) == 0 || m.pp[m.latest] == nil {
			return
		}
		mdl.update(m)
		mdl.report()
	}
}

func run(p *pkg, rslt chan *pkg) {
	rr, err := p.tp.Run()
	p.runResult = &runResult{Results: rr, err: err}
	rslt <- p
}

// reporter implements view.Reporter.
type reporter struct {
	flags view.RprtMask
	ll    []string
	mask  map[uint]view.LineMask
	lst   func(int)
}

// Clearing indicates if all lines not set by this reporter's For
// function should be cleared or not.
func (l *reporter) Flags() view.RprtMask { return l.flags }

// For expects the view's reporting component and a callback to which
// the updated lines can be provided to.
func (l *reporter) For(_ lines.Componenter, line func(uint, string)) {
	for idx, content := range l.ll {
		line(uint(idx), content)
	}
}

// Mask returns for given index special formatting directives.
func (l *reporter) LineMask(idx uint) view.LineMask { return l.mask[idx] }

// Listener returns the callback which is informed about user selections
// of lines by providing the index of the selected line.
func (l *reporter) Listener() func(int) { return l.lst }

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
