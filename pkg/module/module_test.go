// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"testing"
	"time"

	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/fx"
)

func TestModuleDefaultsToWorkingDirectory(t *testing.T) {
	fx, expMod := fx.NewDir(t).CWD(), "example.com/test_wd"
	defer fx.Reset()
	fx.MkMod(expMod)
	m := Module{}
	m.Watch()
	defer m.QuitAll()
	if fx.Name != m.Dir {
		t.Errorf("expected module dir %s; got %s", fx.Name, m.Dir)
	}
	if expMod != m.Name() {
		t.Errorf("expected module name %s; got %s", expMod, m.Name())
	}
}

type module struct{ Suite }

func (s *module) SetUp(t *T) { t.Parallel() }

func (s *module) Fails_initial_watcher_registration_if_no_module(t *T) {
	fx := NewFX(t.GoT())
	defer fx.QuitAll()
	_, _, err := fx.Watch()
	t.FatalIfNot(t.True(err != nil))
	t.ErrIs(err, ErrNoModule)
}

func (s *module) Is_not_watched_if_no_initial_watcher(t *T) {
	fx := NewFX(t.GoT())
	t.False(fx.IsWatched())
}

func (s *module) Is_watched_having_registered_initial_watcher(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod)
	defer fx.QuitAll()

	_, _, err := fx.Watch()
	t.FatalOn(err)

	t.True(fx.IsWatched())
}

func (s *module) Reports_dir_after_initial_watcher(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod)
	defer fx.QuitAll()

	_, _, err := fx.Watch()
	t.FatalOn(err)

	t.Eq(fx.FxDir.Name, fx.Dir)
}

func (s *module) Reports_name_after_initial_watcher(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod)
	defer fx.QuitAll()

	_, _, err := fx.Watch()
	t.FatalOn(err)

	t.Eq(FxModuleName, fx.Name())
}

func (s *module) Reports_package_diffs_to_all_watcher(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod | FxTestingPackage)
	defer fx.QuitAll()
	fx.Interval = 1 * time.Millisecond
	diff1, _, err := fx.Watch()
	t.FatalOn(err)
	diff2, _, err := fx.Watch()
	t.FatalOn(err)
	cond := func() func() bool {
		got1, got2 := false, false
		return func() bool {
			select {
			case diff1 := <-diff1:
				got1 = diff1 != nil
			case diff2 := <-diff2:
				got2 = diff2 != nil
			default:
			}
			return got1 && got2
		}
	}
	t.Within(&TimeStepper{}, cond())
}

func (s *module) Closes_diff_channel_if_watcher_quits(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod | FxTestingPackage)
	defer fx.QuitAll()
	fx.Interval = 1 * time.Millisecond

	diff, ID, err := fx.Watch()
	t.FatalOn(err)

	init := <-diff
	t.True(init != nil)
	fx.Quit(ID)
	t.Within(&TimeStepper{}, func() bool {
		select {
		case diff := <-diff:
			return diff == nil
		default:
			return false
		}
	})
}

func (s *module) Ignores_reserves_zero_quitting_for_quit_all(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod | FxTestingPackage)
	defer fx.QuitAll()
	fx.Interval = 1 * time.Millisecond

	_, _, err := fx.Watch()
	t.FatalOn(err)

	fx.Quit(0)
	fx.Quit(500) // coverage
	t.True(fx.IsWatched())
}

func TestModule(t *testing.T) {
	t.Parallel()
	Run(&module{}, t)
}
