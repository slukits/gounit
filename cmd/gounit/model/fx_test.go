// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Module fixture creation.

package model

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/tfs"
)

type FxMask uint64

const (
	// FxMod indicates a go.mod file creation.
	FxMod FxMask = 1 << iota
	// FxTestingPackage indicates the creation of a package with test
	// file.
	FxTestingPackage
	// FxPackage indicates a package creation without test file
	FxPackage
	// FxParsing indicates the creation of a testing package having the
	// two test files with content fxParseA and fxParseB respectively.
	FxParsing
	// FxTidy runs go mod tidy in a created module fixture and caches
	// the result for reuse.
	FxTidy
	// FxSuiteOrder indicates the creation of a package with three test
	// files with different modification times whose suites should be
	// reported in the order FxSuiteA, FxSuiteD, FxSuiteC and FxSuiteB.
	FxSuiteOrder
)

var allFixtureSettings = []FxMask{
	FxMod, FxTestingPackage, FxParsing,
	// Note FxTidy must be processed after FxMod
	FxTidy, FxSuiteOrder,
}

const (
	FxModuleName   = "example.com/gounit/module"
	fmtPackageName = "fx_package_%s"
	fxTestFileName = "fx"
	fxTest         = "import \"testing\"\n\n" +
		"func TestFixture(t *testing.T) {}\n"
	fxSuiteA   = "FxSuiteA"
	fxSuiteB   = "FxSuiteB"
	fxSuiteC   = "FxSuiteC"
	fxSuiteD   = "FxSuiteD"
	fxARunner  = "TestFxSuiteA"
	fxBRunner  = "TestFxSuiteB"
	fxCRunner  = "TestFxSuiteC"
	fxDRunner  = "TestFxSuiteD"
	fxTestA    = "TestFixtureA"
	fxTestB    = "TestFixtureB"
	fxStATest1 = "Suite_test1"
	fxStATest2 = "SuiteA_test2"
	fxStBTest1 = "Suite_test1"
	fxStBTest2 = "SuiteB_test2"
	fxStCTest  = "SuiteC_test"
	fxStDTest  = "SuiteD_test"
	fxCode     = "type T struct {}"
)

var (
	fxParseA = "import (\n" +
		"\t\"testing\"\n\n" +
		"\t. \"github.com/slukits/gounit\"\n" +
		")\n\n" +
		fmt.Sprintf("func %s(t *testing.T) {\n", fxTestA) +
		"\tt.Error(\"fixture test A: failed\")\n}\n\n" +
		fmt.Sprintf("type %s struct{ Suite }\n\n", fxSuiteA) +
		"func (s *FxSuiteA) Init(t *S) {\n" +
		"\tt.Log(\"log: suite a: init\")\n}\n\n" +
		"func (s *FxSuiteA) SetUp(t *T) {}\n\n" +
		"func (s *FxSuiteA) TearDown(t *T) {}\n\n" +
		fmt.Sprintf(
			"func (s *%s) %s(t *T) {\n", fxSuiteA, fxStATest1) +
		"\tt.Log(\"log: suite a: one\")\n" +
		"\tt.Log(\"log: suite a: two\")\n}\n\n" +
		"func (s *FxSuiteA) PublicHelper() {}\n\n" +
		"func (s *FxSuiteA) Finalize(t *S) {}\n\n" +
		fmt.Sprintf("func %s(t *testing.T) { Run(&%s{}, t) }\n\n",
			fxARunner, fxSuiteA) +
		fmt.Sprintf(
			"func (s *%s) %s(t *T) {}\n\n", fxSuiteB, fxStBTest2)

	fxParseB = "import (\n" +
		"\t\"testing\"\n\n" +
		"\t\"github.com/slukits/gounit\"\n" +
		")\n\n" +
		fmt.Sprintf("func %s(t *testing.T) {\n", fxTestB) +
		"\tt.Log(\"log: test b\")\n}\n\n" +
		fmt.Sprintf("func (s *%s) %s(t *gounit.T) {}\n\n",
			fxSuiteA, fxStATest2) +
		"type FxSuiteB struct{ gounit.Suite }\n\n" +
		"func (s *FxSuiteB) Init(t *gounit.S) {\n" +
		"\tt.Log(\"log: suite b: init\")\n}\n\n" +
		"func (s *FxSuiteB) SetUp(t *gounit.T) {}\n\n" +
		"func (s *FxSuiteB) TearDown(t *gounit.T) {}\n\n" +
		fmt.Sprintf("func (s *%s) %s(t *gounit.T) {\n",
			fxSuiteB, fxStBTest1) +
		"\tt.True(false)\n}\n\n" +
		"func (s *FxSuiteB) privateHelper(t *gounit.T) {}\n\n" +
		"func (s *FxSuiteB) Finalize(t *gounit.S) {\n" +
		"\tt.Log(\"log: suite b: finalize\")\n}\n\n" +
		fmt.Sprintf(
			"func %s(t *testing.T) { gounit.Run(&%s{}, t) }\n\n",
			fxBRunner, fxSuiteB,
		)

	fxOrderA = "import (\n" +
		"\t\"testing\"\n\n" +
		"\t. \"github.com/slukits/gounit\"\n" +
		")\n\n" +
		fmt.Sprintf("type %s struct{ Suite }\n\n", fxSuiteA) +
		fmt.Sprintf(
			"func (s *%s) %s(t *T) {\n", fxSuiteA, fxStATest1) +
		"\tt.Log(\"log: suite a: one\")\n" +
		"\tt.Log(\"log: suite a: two\")\n}\n\n" +
		fmt.Sprintf("func %s(t *testing.T) { Run(&%s{}, t) }\n\n",
			fxARunner, fxSuiteA)

	fxOrderBCD = "import (\n" +
		"\t\"testing\"\n\n" +
		"\t. \"github.com/slukits/gounit\"\n" +
		")\n\n" +
		fmt.Sprintf("type %s struct{ Suite }\n\n", fxSuiteB) +
		fmt.Sprintf(
			"func (s *%s) %s(t *T) {}\n", fxSuiteB, fxStBTest1) +
		fmt.Sprintf("func %s(t *testing.T) { Run(&%s{}, t) }\n\n",
			fxBRunner, fxSuiteB) +
		fmt.Sprintf("type %s struct{ Suite }\n\n", fxSuiteD) +
		fmt.Sprintf(
			"func (s *%s) %s(t *T) {}\n", fxSuiteD, fxStDTest) +
		fmt.Sprintf("func %s(t *testing.T) { Run(&%s{}, t) }\n\n",
			fxDRunner, fxSuiteD) +
		fmt.Sprintf("type %s struct{ Suite }\n\n", fxSuiteC) +
		fmt.Sprintf(
			"func (s *%s) %s(t *T) {}\n", fxSuiteC, fxStCTest) +
		fmt.Sprintf("func %s(t *testing.T) { Run(&%s{}, t) }\n\n",
			fxCRunner, fxSuiteC)

	fxOrderB = "import (\n" +
		"\t\"testing\"\n\n" +
		"\t. \"github.com/slukits/gounit\"\n" +
		")\n\n" +
		fmt.Sprintf(
			"func (s *%s) %s(t *T) {}\n", fxSuiteB, fxStBTest2)
)

// ModuleFX wraps a Module instance and augments it with testing
// convenance features.  E.g. Quit which makes sure all watchers are
// quit.
type ModuleFX struct {
	*Sources
	t      *gounit.T
	fxWW   []chan struct{}
	FxDir  *tfs.Dir
	n, tpN int
}

func NewFX(t *gounit.T) *ModuleFX {
	d := t.FS().Tmp()
	return &ModuleFX{
		Sources: &Sources{Dir: d.Path()}, t: t, FxDir: d}
}

func (x *ModuleFX) Set(ff FxMask) *ModuleFX {
	for _, f := range allFixtureSettings {
		if ff&f != f {
			continue
		}
		x.set(f)
	}
	return x
}

func (x *ModuleFX) set(f FxMask) {
	switch f {
	case FxMod:
		x.FxDir.MkMod(FxModuleName)
	case FxTestingPackage:
		packageName := x.newTestingPackageName()
		pkgDir, _ := x.FxDir.Mk(packageName)
		pkgDir.MkPkgTest(fxTestFileName, []byte(fxTest))
		pkgDir.MkPkgFile(fxTestFileName, []byte(fxCode))
	case FxPackage:
		packageName := x.newPackageName()
		pkgDir, _ := x.FxDir.Mk(packageName)
		pkgDir.MkPkgFile(fxTestFileName, []byte(fxCode))
	case FxParsing:
		packageName := x.newTestingPackageName()
		pkgDir, _ := x.FxDir.Mk(packageName)
		pkgDir.MkPkgTest(
			fmt.Sprintf("%sa", fxTestFileName), []byte(fxParseA))
		pkgDir.MkPkgTest(
			fmt.Sprintf("%sb", fxTestFileName), []byte(fxParseB))
	case FxSuiteOrder:
		packageName := x.newTestingPackageName()
		pkgDir, _ := x.FxDir.Mk(packageName)
		pkgDir.MkPkgTest(
			fmt.Sprintf("%sa", fxTestFileName), []byte(fxOrderA))
		time.Sleep(5 * time.Millisecond)
		pkgDir.MkPkgTest(
			fmt.Sprintf("%sbcd", fxTestFileName), []byte(fxOrderBCD))
		time.Sleep(5 * time.Millisecond)
		pkgDir.MkPkgTest(
			fmt.Sprintf("%sb", fxTestFileName), []byte(fxOrderB))
	case FxTidy:
		x.FxDir.MkTidy()
	}
}

// IsTesting returns true iff given package name is a testing package
// name created by this fixture.
func (x *ModuleFX) IsTesting(pkg string) bool {
	for i := 1; i <= x.tpN; i++ {
		if x.testingPackageNameOf(i) != pkg {
			continue
		}
		return true
	}
	return false
}

// ForTesting calls back for each created testing package fixture by
// this module fixture.  NOTE this package is not guaranteed to exist.
func (x *ModuleFX) ForTesting(cb func(string) (stop bool)) {
	for i := 1; i <= x.tpN; i++ {
		if cb(x.testingPackageNameOf(i)) {
			return
		}
	}
}

func (x *ModuleFX) RM(packageName string) {
	err := os.RemoveAll(filepath.Join(x.FxDir.Path(), packageName))
	if err != nil {
		x.t.Fatalf("couldn't remove package '%s': %v",
			packageName, err)
	}
}

func (x *ModuleFX) RMFile(relFile string) {
	err := os.Remove(filepath.Join(x.FxDir.Path(), relFile))
	if err != nil {
		x.t.Fatalf("couldn't remove file '%s': %v",
			filepath.Join(x.FxDir.Path(), relFile), err)
	}
}

func (x *ModuleFX) newTestingPackageName() string {
	x.tpN++
	return x.testingPackageNameOf(x.tpN)
}

func (x *ModuleFX) testingPackageNameOf(n int) string {
	return fmt.Sprintf(fmtPackageName, fmt.Sprintf("t%d", n))
}

func (x *ModuleFX) newPackageName() string {
	x.n++
	return x.packageNameOf(x.n)
}

func (x *ModuleFX) packageNameOf(n int) string {
	return fmt.Sprintf(fmtPackageName, strconv.Itoa(n))
}

// createFixturePkg is a cheaper version of creating a TestingPackage
// without a go.mod and go.sum file, i.e. running its tests will fail
// but it suffices for evaluating tests-parsing results and source
// statistics.
func createFixturePkg(t *gounit.T, testDataDir string) *TestingPackage {
	testData, _ := t.FS().Data()
	tmp := t.FS().Tmp()
	testData.Child(testDataDir).Copy(tmp)
	pkgStats, ok := newTestingPackageStat(
		tmp.Child(testDataDir).Path())
	if !ok {
		t.Fatalf("failed to obtain package stats of %s",
			tmp.Child(testDataDir).Path())
	}
	tt, err := pkgStats.loadTestFiles()
	t.FatalOn(err)
	return &TestingPackage{abs: pkgStats.abs, files: tt}
}
