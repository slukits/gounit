// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"testing"

	. "github.com/slukits/gounit"
)

type Util struct{ Suite }

func (s *Util) SetUp(t *T) { t.Parallel() }

func (s *Util) Reports_module_path_and_name_in_given_directory(t *T) {
	fx, exp := t.FS().Tmp(), "github.com/slukits/test"
	fx.MkMod(exp)
	nested, _ := fx.MkTmp("dirA", "dirB", "dirC")

	gotDir, gotName, err := findModule(nested.Path())
	t.FatalOn(err)

	t.Eq(fx.Path(), gotDir)
	t.Eq(exp, gotName)
}

func (s *Util) Reports_no_module_if_no_go_mod_in_given_path(t *T) {
	nested, _ := t.FS().Tmp().Mk("dirA", "dirB", "dirC")

	_, _, err := findModule(nested.Path())

	t.FatalIfNot(t.True(err != nil))
	t.ErrIs(err, ErrNoModule)
}

func TestUtil(t *testing.T) {
	t.Parallel()
	Run(&Util{}, t)
}
