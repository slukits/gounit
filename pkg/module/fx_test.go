// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Module fixture creation.

package module

import (
	"testing"

	"github.com/slukits/gounit/pkg/fx"
)

type FxMask uint64

const (
	FxMod FxMask = 1 << iota
	FxTestingPackage
)

var allFixtureSettings = []FxMask{FxMod, FxTestingPackage}

const (
	FxModuleName   = "example.com/gounit/module"
	FxPackageName  = "fx_package"
	FxTestFileName = "fx"
	FxTest         = "import \"testing\"\n\n" +
		"func TestFixture(t *testing.T) {\n" +
		"}\n"
)

// ModuleFX wraps a Module instance and augments it with testing
// convenance features.  E.g. Quit which makes sure all watchers are
// quit.
type ModuleFX struct {
	*Module
	fxWW  []chan struct{}
	FxDir *fx.Dir
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
		x.FxDir.MkPath(FxPackageName)
		x.FxDir.MkPkgTest(FxPackageName, FxTestFileName, FxTest)
	}
}
