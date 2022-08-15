// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/fx"
)

type Util struct{ Suite }

func (s *Util) SetUp(t *T) { t.Parallel() }

func (s *Util) Reports_module_path_and_name_in_given_directory(t *T) {
	exp := "github.com/slukits/test"
	fx := fx.NewDir(t.GoT()).MkMod(exp)
	_, path := fx.MkPath("dirA", "dirB", "dirC")

	gotDir, gotName, err := findModule(path)
	t.FatalOn(err)

	t.Eq(fx.Name, gotDir)
	t.Eq(exp, gotName)
}

func (s *Util) Reports_no_module_if_no_go_mod_in_given_path(t *T) {
	_, path := fx.NewDir(t.GoT()).MkPath("dirA", "dirB", "dirC")

	_, _, err := findModule(path)

	t.FatalIfNot(t.True(err != nil))
	t.ErrIs(err, ErrNoModule)
}

func TestUtil(t *testing.T) {
	t.Parallel()
	Run(&Util{}, t)
}
