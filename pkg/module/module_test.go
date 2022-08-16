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

func (s *module) Is_unwatched_if_all_watcher_quit(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod | FxTestingPackage)
	defer fx.QuitAll()
	fx.Interval = 1 * time.Millisecond
	_, ID1, err := fx.Watch()
	t.FatalOn(err)
	_, ID2, err := fx.Watch()
	t.FatalOn(err)
	t.True(fx.IsWatched())

	fx.Quit(ID1)
	fx.Quit(ID2)

	t.Within(&TimeStepper{}, func() bool { return !fx.IsWatched() })
}

func (s *module) Reserves_zero_quitting_for_quit_all(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod | FxTestingPackage)
	fx.Interval = 1 * time.Millisecond

	_, _, err := fx.Watch()
	t.FatalOn(err)

	fx.Quit(0)
	fx.Quit(500) // coverage
	t.True(fx.IsWatched())

	fx.QuitAll()
	t.False(fx.IsWatched())
}

func (s *module) Reports_initially_all_testing_packages(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod | FxTestingPackage)
	fx.Set(FxPackage | FxTestingPackage)
	fx.Interval = 1 * time.Millisecond
	defer fx.QuitAll()

	diff, _, err := fx.Watch()
	t.FatalOn(err)
	var init *PackagesDiff
	select {
	case init = <-diff:
		t.FatalIfNot(t.True(init != nil))
	case <-t.Timeout(0):
		t.Fatal("initial diff-report timed out")
	}
	gotN := 0

	init.For(func(tp *TestingPackage) (stop bool) {
		t.True(fx.IsTesting(tp.Name()))
		gotN++
		return false
	})
	t.Eq(2, gotN)
}

func testWatcher(diff <-chan *PackagesDiff) (initOnly chan bool) {
	initOnly, state := make(chan bool), 0
	go func() {
		for {
			select {
			case diff := <-diff:
				if diff != nil && state == 0 {
					state = 1
					continue
				}
				if diff != nil && state == 1 {
					state = -1
				}
			case respond := <-initOnly:
				if !respond {
					close(initOnly)
					return
				}
				initOnly <- state == 1
			}
		}
	}()
	return initOnly
}

func (s *module) Reports_all_testing_packages_to_new_watcher(t *T) {
	fx := NewFX(t.GoT()).Set(FxMod | FxTestingPackage)
	fx.Set(FxPackage | FxTestingPackage)
	fx.Interval = 1 * time.Millisecond
	defer fx.QuitAll()
	diff1, _, err := fx.Watch()
	t.FatalOn(err)
	diff2, _, err := fx.Watch()
	t.FatalOn(err)
	initOnly1, initOnly2 := testWatcher(diff1), testWatcher(diff2)

	t.Within(&TimeStepper{}, func() bool {
		initOnly1 <- true
		io1 := <-initOnly1
		initOnly2 <- true
		io2 := <-initOnly2
		return io1 && io2
	})
	initOnly1 <- false
	initOnly2 <- false
}

type dbg struct{ Suite }

func (s *dbg) Dbg(t *T) {
}

func TestDBG(t *testing.T) { Run(&dbg{}, t) }
func TestModule(t *testing.T) {
	t.Parallel()
	Run(&module{}, t)
}
