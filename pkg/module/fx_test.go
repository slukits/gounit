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
	FxMod FxMask = 1 << iota
	FxTestingPackage
	FxPackage
)

var allFixtureSettings = []FxMask{FxMod, FxTestingPackage}

const (
	FxModuleName   = "example.com/gounit/module"
	fmtPackageName = "fx_package_%s"
	fxTestFileName = "fx"
	fxTest         = "import \"testing\"\n\n" +
		"func TestFixture(t *testing.T) {}\n"
	fxCode = "type T struct {}"
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
