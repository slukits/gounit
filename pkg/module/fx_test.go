// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Module fixture creation.

package module

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/slukits/gounit/pkg/fx"
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
)

var allFixtureSettings = []FxMask{
	FxMod, FxTestingPackage, FxParsing}

const (
	FxModuleName   = "example.com/gounit/module"
	fmtPackageName = "fx_package_%s"
	fxTestFileName = "fx"
	fxTest         = "import \"testing\"\n\n" +
		"func TestFixture(t *testing.T) {}\n"
	fxSuiteA   = "FxSuiteA"
	fxSuiteB   = "FxSuiteB"
	fxARunner  = "TestFxSuiteA"
	fxBRunner  = "TestFxSuiteB"
	fxTestA    = "TestFixtureA"
	fxTestB    = "TestFixtureB"
	fxStATest1 = "Suite_test1"
	fxStATest2 = "SuiteA_test2"
	fxStBTest1 = "Suite_test1"
	fxStBTest2 = "SuiteB_test2"
	fxCode     = "type T struct {}"
)

var (
	fxParseA = "import (\n" +
		"\t\"testing\"\n\n" +
		"\t. \"github.com/slukits/gounit\"\n" +
		")\n\n" +
		fmt.Sprintf("func %s(t *testing.T) {}\n\n", fxTestA) +
		fmt.Sprintf("type %s struct{ Suite }\n\n", fxSuiteA) +
		"func (s *FxSuiteA) Init(t *S) {}\n\n" +
		"func (s *FxSuiteA) SetUp(t *T) {}\n\n" +
		"func (s *FxSuiteA) TearDown(t *T) {}\n\n" +
		fmt.Sprintf(
			"func (s *%s) %s(t *T) {}\n\n", fxSuiteA, fxStATest1) +
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
		fmt.Sprintf("func %s(t *testing.T) {}\n\n", fxTestB) +
		fmt.Sprintf("func (s *%s) %s(t *gounit.T) {}\n\n",
			fxSuiteA, fxStATest2) +
		"type FxSuiteB struct{ gounit.Suite }\n\n" +
		"func (s *FxSuiteB) Init(t *gounit.S) {}\n\n" +
		"func (s *FxSuiteB) SetUp(t *gounit.T) {}\n\n" +
		"func (s *FxSuiteB) TearDown(t *gounit.T) {}\n\n" +
		fmt.Sprintf("func (s *%s) %s(t *gounit.T) {}\n\n",
			fxSuiteB, fxStBTest1) +
		"func (s *FxSuiteB) privateHelper(t *gounit.T) {}\n\n" +
		"func (s *FxSuiteB) Finalize(t *gounit.S) {}\n\n" +
		fmt.Sprintf(
			"func %s(t *testing.T) { gounit.Run(&%s{}, t) }\n\n",
			fxBRunner, fxSuiteB,
		)
)

// ModuleFX wraps a Module instance and augments it with testing
// convenance features.  E.g. Quit which makes sure all watchers are
// quit.
type ModuleFX struct {
	*Module
	fxWW   []chan struct{}
	FxDir  *fx.Dir
	n, tpN int
}

func NewFX(t *testing.T) *ModuleFX {
	d := fx.NewDir(t)
	return &ModuleFX{Module: &Module{Dir: d.Name}, FxDir: d}
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
		x.FxDir.MkPath(packageName)
		x.FxDir.MkPkgTest(packageName, fxTestFileName, fxTest)
		x.FxDir.MkPkgFile(packageName, fxTestFileName, fxCode)
	case FxPackage:
		packageName := x.newPackageName()
		x.FxDir.MkPath(packageName)
		x.FxDir.MkPkgFile(packageName, fxTestFileName, fxCode)
	case FxParsing:
		packageName := x.newTestingPackageName()
		x.FxDir.MkPath(packageName)
		x.FxDir.MkPkgTest(packageName,
			fmt.Sprintf("%sa", fxTestFileName), fxParseA)
		x.FxDir.MkPkgTest(packageName,
			fmt.Sprintf("%sb", fxTestFileName), fxParseB)
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
	err := os.RemoveAll(filepath.Join(x.FxDir.Name, packageName))
	if err != nil {
		x.FxDir.T.Fatalf("couldn't remove package '%s': %v",
			packageName, err)
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
