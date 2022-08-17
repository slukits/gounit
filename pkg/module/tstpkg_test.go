// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

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

func (fx *fxSet) TestingPackage(t *T) *TestingPackage {
	diff, _, err := fx.Get(t).(*ModuleFX).Watch()
	t.FatalOn(err)
	select {
	case diff := <-diff:
		var pkg *TestingPackage
		diff.For(func(tp *TestingPackage) (stop bool) {
			pkg = tp
			return true
		})
		t.FatalIfNot(t.True(pkg != nil))
		return pkg
	case <-t.Timeout(0):
		t.Fatal("initial diff timed out")
	}
	return nil
}

func (s *Package) SetUp(t *T) {
	t.Parallel()
	fx := NewFX(t.GoT()).Set(FxMod | FxParsing)
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

func TestPackage(t *testing.T) {
	t.Parallel()
	Run(&Package{}, t)
}
