// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"testing"
	"time"

	. "github.com/slukits/gounit"
)

type Package struct {
	Suite
	fx fxSet
}

type fxSet struct{ Fixtures }

func (fx *fxSet) initTestingPackage(
	t *T, c <-chan *PackagesDiff,
) *TestingPackage {
	select {
	case diff := <-c:
		var pkg *TestingPackage
		diff.For(func(tp *TestingPackage) (stop bool) {
			pkg = tp
			return true
		})
		t.FatalIfNot(t.True(pkg != nil))
		return pkg
	case <-t.Timeout(30 * time.Millisecond):
		t.Fatal("initial diff timed out")
	}
	return nil
}

func (fx *fxSet) TestingPackage(t *T) *TestingPackage {
	diff, _, err := fx.Get(t).(*ModuleFX).Watch()
	t.FatalOn(err)
	return fx.initTestingPackage(t, diff)
}

func (fx *fxSet) Module(t *T) *ModuleFX { return fx.Get(t).(*ModuleFX) }

func (s *Package) SetUp(t *T) {
	t.Parallel()
	fx := NewFX(t).Set(FxMod | FxParsing)
	fx.Interval = 1 * time.Millisecond
	s.fx.Set(t, fx)
}

func (s *Package) TearDown(t *T) {
	s.fx.Del(t).(*ModuleFX).QuitAll()
}

func (s *Package) Reports_tests(t *T) {
	exp, n := map[string]bool{fxTestA: true, fxTestB: true}, 0
	pkg := s.fx.TestingPackage(t)

	pkg.ForTest(func(tst *Test) {
		n++
		t.True(exp[tst.Name()])
	})

	t.Eq(len(exp), n)
}

func (s *Package) Reports_suites(t *T) {
	exp, n := map[string]bool{fxSuiteA: true, fxSuiteB: true}, 0
	pkg := s.fx.TestingPackage(t)

	pkg.ForSuite(func(st *TestSuite) {
		n++
		t.True(exp[st.Name()])
	})

	t.Eq(len(exp), n)
}

func (s *Package) Reports_the_number_of_suites(t *T) {
	pkg := s.fx.TestingPackage(t)
	t.Eq(2, pkg.LenSuites())
}

func (s *Package) Reports_suites_ordered(t *T) {
	pkg, n := s.fx.TestingPackage(t), 0
	exp := []string{"FxSuiteA", "FxSuiteB"}

	pkg.ForSortedSuite(func(st *TestSuite) {
		t.Eq(exp[n], st.Name())
		n++
	})

	t.Eq(2, n)
}

func (s *Package) Provides_suite_by_name(t *T) {
	pkg := s.fx.TestingPackage(t)
	t.Eq("FxSuiteA", pkg.Suite("FxSuiteA").Name())
}

func (s *Package) Provides_last_parsed_suite(t *T) {
	pkg := s.fx.TestingPackage(t)
	t.Eq("FxSuiteB", pkg.LastSuite().Name())
}

func (s *Package) Reports_suite_tests(t *T) {
	exp := map[string]map[string]bool{
		fxSuiteA: {fxStATest1: true, fxStATest2: true},
		fxSuiteB: {fxStBTest1: true, fxStBTest2: true},
	}
	pkg := s.fx.TestingPackage(t)

	pkg.ForSuite(func(st *TestSuite) {
		n := 0
		st.ForTest(func(tst *Test) {
			n++
			sn, tn := st.Name(), tst.Name()
			t.True(exp[sn][tn])
		})
		t.Eq(len(exp[st.Name()]), n)
	})
}

func (s *Package) Reports_suite_runner(t *T) {
	exp := map[string]string{fxSuiteA: fxARunner, fxSuiteB: fxBRunner}
	pkg := s.fx.TestingPackage(t)

	pkg.ForSuite(func(st *TestSuite) {
		t.Eq(exp[st.Name()], st.Runner())
	})
}

func (s *Package) Reports_shell_exit_error_of_tests_run(t *T) {
	pkg := s.fx.TestingPackage(t)

	rslt, err := pkg.Run(0)
	t.FatalOn(err)

	t.True(rslt.HasErr())
	t.Contains(rslt.Err(), StdErr)
}

func (s *Package) fxSuiteOrder(t *T) *TestingPackage {
	fx := NewFX(t).Set(FxMod | FxSuiteOrder)
	fx.Interval = 1 * time.Millisecond
	diff, _, err := fx.Watch()
	t.FatalOn(err)
	return s.fx.initTestingPackage(t, diff)
}

func (s *Package) Reports_suite_of_most_recent_test_file_last(t *T) {
	fx, idx := s.fxSuiteOrder(t), 0
	exp := []string{fxSuiteA, fxSuiteD, fxSuiteC, fxSuiteB}
	fx.ForSuite(func(ts *TestSuite) {
		t.Eq(exp[idx], ts.Name())
		idx++
	})
}

func (s *Package) Has_initially_no_source_stats(t *T) {
	fx := createSourceFixture(t)
	t.Not.True(fx.HasSrcStats())
}

func (s *Package) Has_source_stats_after_they_were_requested(t *T) {
	fx := createSourceFixture(t)
	fx.SrcStats()
	t.True(fx.HasSrcStats())
}

func (s *Package) Has_no_source_stats_after_resetting_them(t *T) {
	fx := createSourceFixture(t)
	fx.SrcStats()
	t.True(fx.HasSrcStats())
	fx.ResetSrcStats()
	t.Not.True(fx.HasSrcStats())
}

func (s *Package) Reported_tests_stringers_are_humanized_names(t *T) {
	fx, count := createFixturePkg(t, "humanizefx"), 0
	exp := map[string]bool{
		"broken END": true, "HTTP acronym": true, "HTTP2 acronym": true,
		"camel snake mixed": true, "HTTP2 snake case": true,
		"increment ID": true,
	}
	fx.ForTest(func(tst *Test) {
		t.True(exp[tst.String()])
		count++
	})
	t.Eq(6, count)
}

func TestPackage(t *testing.T) {
	t.Parallel()
	Run(&Package{}, t)
}
