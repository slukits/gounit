// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/ints"
)

type features struct{ Suite }

func (s *features) Has_by_default_only_quit_registered(t *T) {
	exp := (&ints.Set{}).Add(int(FtQuit))
	t.True(exp.Eq(DefaultFeatures.Registered()))
}

func (s *features) Of_defaults_may_not_be_changed(t *T) {
	exp := (&ints.Set{}).Add(int(FtQuit))
	DefaultFeatures.Add(FtUp, 'k', tcell.KeyUp, 0)
	t.True(exp.Eq(DefaultFeatures.Registered()))
	DefaultFeatures.Del(FtQuit)
	t.True(exp.Eq(DefaultFeatures.Registered()))
}

func TestFeatures(t *testing.T) {
	t.Parallel()
	Run(&features{}, t)
}
