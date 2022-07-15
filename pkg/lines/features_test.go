// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines_test

import (
	"testing"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/gounit/pkg/lines"
	"github.com/slukits/ints"
)

type Features struct{ Suite }

func (s *Features) Has_by_default_only_quit_registered(t *T) {
	exp := (&ints.Set{}).Add(int(lines.FtQuit))
	t.True(exp.Eq(lines.DefaultFeatures.Registered()))
}

func (s *Features) Of_defaults_may_not_be_changed(t *T) {
	exp := (&ints.Set{}).Add(int(lines.FtQuit))
	lines.DefaultFeatures.Add(lines.FtUp, 'k', tcell.KeyUp, 0)
	t.True(exp.Eq(lines.DefaultFeatures.Registered()))
	lines.DefaultFeatures.Del(lines.FtQuit)
	t.True(exp.Eq(lines.DefaultFeatures.Registered()))
}

func TestFeatures(t *testing.T) {
	t.Parallel()
	Run(&Features{}, t)
}
