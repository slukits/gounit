// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"testing"
	"time"

	. "github.com/slukits/gounit"
)

type Packages struct{ Suite }

func (s *Packages) Are_reported_descending_by_modification_date(t *T) {
	fx := NewFX(t).Set(FxMod | FxTestingPackage)
	// create different modification times
	time.Sleep(5 * time.Millisecond)
	defer fx.QuitAll()
	fx.Set(FxPackage | FxTestingPackage)
	time.Sleep(5 * time.Millisecond)
	fx.Set(FxTestingPackage)
	fx.Interval = 10 * time.Millisecond
	diff, _, err := fx.Watch()
	t.FatalOn(err)

	var last *TestingPackage
	select {
	case diff := <-diff:
		diff.For(func(tp *TestingPackage) (stop bool) {
			if last == nil {
				last = tp
				return
			}
			t.True(tp.ModTime.Before(last.ModTime))
			last = tp
			return
		})
	case <-t.Timeout(30 * time.Millisecond):
		t.Error("initial diff report timed out")
	}
}

func TestPackages(t *testing.T) {
	Run(&Packages{}, t)
}
