// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"testing"
	"time"

	. "github.com/slukits/gounit"
)

func TestModuleDefaultsToWorkingDirectory(t *testing.T) {
	fx, expMod := NewT(t).FS().Tmp(), "example.com/test_wd"
	undo := fx.CWD()
	defer undo()
	fx.MkMod(expMod)
	m := Sources{}
	m.Watch()
	defer m.QuitAll()
	if fx.Path() != m.Dir {
		t.Errorf("expected module dir %s; got %s", fx.Path(), m.Dir)
	}
	if expMod != m.ModuleName() {
		t.Errorf("expected module name %s; got %s", expMod, m.ModuleName())
	}
}

type source struct{ Suite }

func (s *source) SetUp(t *T) { t.Parallel() }

func (s *source) Fails_initial_watcher_registration_if_no_module(t *T) {
	fx := NewFX(t)
	defer fx.QuitAll()
	_, _, err := fx.Watch()
	t.FatalIfNot(t.True(err != nil))
	t.ErrIs(err, ErrNoModule)
}

func (s *source) Is_not_watched_if_no_initial_watcher(t *T) {
	fx := NewFX(t)
	t.Not.True(fx.IsWatched())
}

func (s *source) Is_watched_having_registered_initial_watcher(t *T) {
	fx := NewFX(t).Set(FxMod)
	defer fx.QuitAll()

	_, _, err := fx.Watch()
	t.FatalOn(err)

	t.True(fx.IsWatched())
}

func (s *source) Reports_dir_after_initial_watcher(t *T) {
	fx := NewFX(t).Set(FxMod)
	defer fx.QuitAll()

	_, _, err := fx.Watch()
	t.FatalOn(err)

	t.Eq(fx.FxDir.Path(), fx.Dir)
}

func (s *source) Reports_name_after_initial_watcher(t *T) {
	fx := NewFX(t).Set(FxMod)
	defer fx.QuitAll()

	_, _, err := fx.Watch()
	t.FatalOn(err)

	t.Eq(FxModuleName, fx.ModuleName())
}

func (s *source) Reports_package_diffs_to_all_watcher(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
	defer fx.QuitAll()
	fx.Interval = 1 * time.Millisecond

	diff1, _, err := fx.Watch()
	t.FatalOn(err)
	diff2, _, err := fx.Watch()
	t.FatalOn(err)

	t.Within(&TimeStepper{}, func() func() bool {
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
	}())
}

func (s *source) Closes_diff_channel_if_watcher_quits(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
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

func (s *source) Is_unwatched_if_all_watcher_quit(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
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

func (s *source) Reserves_zero_quitting_for_quit_all(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
	fx.Interval = 1 * time.Millisecond

	_, _, err := fx.Watch()
	t.FatalOn(err)

	fx.Quit(0)
	fx.Quit(500) // coverage
	t.True(fx.IsWatched())

	fx.QuitAll()
	t.Not.True(fx.IsWatched())
}

func (s *source) Reports_initially_all_testing_packages(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
	fx.Set(FxPackage | FxTestingPackage)
	fx.Interval = 1 * time.Millisecond
	defer fx.QuitAll()

	diff, _, err := fx.Watch()
	t.FatalOn(err)

	t.Within(&TimeStepper{}, func() bool {
		select {
		default:
			return false
		case diff := <-diff:
			if diff == nil {
				t.Fatal("expected initial diff")
			}
			gotN := 0
			diff.For(func(tp *TestingPackage) (stop bool) {
				t.True(fx.IsTesting(tp.Name()))
				gotN++
				return false
			})
			t.Eq(2, gotN)
			return true
		}
	})
}

func testWatcher(
	t *T, fx *ModuleFX, diff <-chan *PackagesDiff,
) (initOnly chan bool) {
	initOnly, state := make(chan bool), 0
	go func() {
		for {
			select {
			case diff := <-diff:
				if diff != nil && state == 0 {
					// ensure all packages are reported initially
					gotN := 0
					diff.For(func(tp *TestingPackage) (stop bool) {
						t.True(fx.IsTesting(tp.Name()))
						gotN++
						return false
					})
					t.Eq(2, gotN)
					state = 1
					continue
				}
				if diff != nil && state == 1 {
					// ensure they are reported only once
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

func (s *source) Reports_all_testing_packages_to_new_watcher(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
	fx.Set(FxPackage | FxTestingPackage)
	fx.Interval = 1 * time.Millisecond
	defer fx.QuitAll()

	diff1, _, err := fx.Watch()
	t.FatalOn(err)
	initOnly1 := testWatcher(t, fx, diff1)
	t.Within(&TimeStepper{}, func() bool {
		initOnly1 <- true
		return <-initOnly1
	})
	diff2, _, err := fx.Watch()
	t.FatalOn(err)
	initOnly2 := testWatcher(t, fx, diff2)

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

func (s *source) Reports_added_package_to_registered_watcher(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
	fx.Interval = 1 * time.Millisecond
	defer fx.QuitAll()

	diff, _, err := fx.Watch()
	t.FatalOn(err)

	init, n := (*PackagesDiff)(nil), 0
	select {
	case init = <-diff:
	case <-t.Timeout(30 * time.Millisecond):
		t.Fatal("expected initial diff-report")
	}
	init.For(func(tp *TestingPackage) (stop bool) {
		n++
		t.True(fx.IsTesting(tp.Name()))
		return false
	})
	t.Eq(1, n)

	fx.Set(FxPackage | FxTestingPackage)
	add, n := (*PackagesDiff)(nil), 0
	select {
	case add = <-diff:
	case <-t.Timeout(30 * time.Millisecond):
		t.Fatal("expected diff-report for added package")
	}
	add.For(func(tp *TestingPackage) (stop bool) {
		n++
		t.True(fx.IsTesting(tp.Name()))
		return false
	})
	t.Eq(1, n)
}

func (s *source) Reports_deleted_package_to_registered_watcher(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
	fx.Set(FxPackage | FxTestingPackage)
	fx.Interval = 1 * time.Millisecond
	defer fx.QuitAll()

	diff, _, err := fx.Watch()
	t.FatalOn(err)
	select {
	case <-diff:
	case <-t.Timeout(30 * time.Millisecond):
		t.Fatal("expected initial diff-report")
	}

	var delPack string
	fx.ForTesting(func(s string) (stop bool) {
		fx.RM(s)
		delPack = s
		return true
	})
	del, n := (*PackagesDiff)(nil), 0
	select {
	case del = <-diff:
	case <-t.Timeout(30 * time.Millisecond):
		t.Fatal("expected diff-report for deleted package")
	}
	del.ForDel(func(tp *TestingPackage) (stop bool) {
		n++
		t.Eq(delPack, tp.Name())
		return false
	})
	t.Eq(1, n)
}

func TestSource(t *testing.T) {
	t.Parallel()
	Run(&source{}, t)
}
