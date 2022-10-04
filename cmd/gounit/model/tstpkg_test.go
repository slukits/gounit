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

func (s *Package) Reports_suits(t *T) {
	exp, n := map[string]bool{fxSuiteA: true, fxSuiteB: true}, 0
	pkg := s.fx.TestingPackage(t)

	pkg.ForSuite(func(st *TestSuite) {
		n++
		t.True(exp[st.Name()])
	})

	t.Eq(len(exp), n)
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

func TestPackage(t *testing.T) {
	t.Parallel()
	Run(&Package{}, t)
}

type PkgTestRun struct {
	Suite
	rslt *Results
	pkg  *TestingPackage
}

func (s *PkgTestRun) Init(t *S) {
	_t := NewT(t.GoT())
	fx := NewFX(_t).Set(FxMod | FxParsing | FxTidy)
	fx.Interval = 1 * time.Millisecond
	diff, _, err := fx.Watch()
	t.FatalOn(err)
	var pkg *TestingPackage
	select {
	case diff := <-diff:
		diff.For(func(tp *TestingPackage) (stop bool) {
			pkg = tp
			return true
		})
		_t.FatalIfNot(_t.True(pkg != nil))
	case <-_t.Timeout(30 * time.Millisecond):
		t.Fatal("initial diff timed out")
	}
	rslt, err := pkg.Run(0)
	t.FatalOn(err)
	_t.FatalIfNot(_t.True(rslt.Err() == ""))

	s.pkg = pkg
	s.rslt = rslt
}

func (s *PkgTestRun) Reports_a_result_for_each_test(t *T) {
	s.pkg.ForTest(func(tst *Test) {
		t.True(s.rslt.OfTest(tst) != nil)
	})
}

func (s *PkgTestRun) Reports_a_result_for_each_suite(t *T) {
	s.pkg.ForSuite(func(tst *TestSuite) {
		t.True(s.rslt.OfSuite(tst) != nil)
	})
}

func (s *PkgTestRun) Reports_results_for_suite_tests(t *T) {
	s.pkg.ForSuite(func(st *TestSuite) {
		st.ForTest(func(tst *Test) {
			t.True(s.rslt.OfSuite(st).OfTest(tst) != nil)
		})
	})
}

func (s *PkgTestRun) Reports_failing_and_passing_of_tests(t *T) {
	s.pkg.ForTest(func(tst *Test) {
		switch tst.Name() {
		case fxTestA:
			t.Not.True(s.rslt.OfTest(tst).Passed)
		case fxTestB:
			t.True(s.rslt.OfTest(tst).Passed)
		}
	})
}

func (s *PkgTestRun) Reports_failing_and_passing_of_suite_tests(
	t *T,
) {
	s.pkg.ForSuite(func(ts *TestSuite) {
		sr := s.rslt.OfSuite(ts)
		ts.ForTest(func(tst *Test) {
			if ts.Name() == fxSuiteB && tst.Name() == fxStBTest1 {
				t.Not.True(sr.OfTest(tst).Passed)
				return
			}
			t.True(sr.OfTest(tst).Passed)
		})
	})
}

func (s *PkgTestRun) Reports_the_number_of_tests(t *T) {
	t.Eq(4, s.rslt.Len())
}

func (s *PkgTestRun) Reports_number_of_suite_tests(t *T) {
	s.pkg.ForSuite(func(ts *TestSuite) {
		sr := s.rslt.OfSuite(ts)
		t.Eq(2, sr.Len())
	})
}

func (s *PkgTestRun) Reports_number_of_failed_suite_tests(t *T) {
	s.pkg.ForSuite(func(ts *TestSuite) {
		sr := s.rslt.OfSuite(ts)
		switch ts.Name() {
		case fxSuiteA:
			t.Eq(0, sr.LenFailed())
		case fxSuiteB:
			t.Eq(1, sr.LenFailed())
		}
	})
}

func (s *PkgTestRun) Reports_test_logs(t *T) {
	s.pkg.ForTest(func(tst *Test) {
		if tst.Name() != fxTestB {
			return
		}
		t.True(len(s.rslt.OfTest(tst).Output) > 0)
	})
}

func (s *PkgTestRun) Reports_suite_test_logs(t *T) {
	s.pkg.ForSuite(func(st *TestSuite) {
		if st.Name() != fxSuiteA {
			return
		}
		sr := s.rslt.OfSuite(st)
		st.ForTest(func(tst *Test) {
			if tst.Name() != fxStATest1 {
				return
			}
			t.Eq(2, len(sr.OfTest(tst).Output))
		})
	})
}

func (s *PkgTestRun) Reports_suit_init_finalize_logs(t *T) {
	s.pkg.ForSuite(func(st *TestSuite) {
		sr := s.rslt.OfSuite(st)
		switch st.Name() {
		case fxSuiteA:
			t.Eq(1, len(sr.InitOut))
		case fxSuiteB:
			t.Eq(1, len(sr.FinalizeOut))
		}
	})
}

func TestPkgTestRun(t *testing.T) {
	t.Parallel()
	Run(&PkgTestRun{}, t)
}
