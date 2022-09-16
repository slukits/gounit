// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type Report struct {
	Suite
	Fixtures
}

func (s *Report) SetUp(t *T) { t.Parallel() }

func (s *Report) TearDown(t *T) {
	fx := s.Get(t)
	if fx == nil {
		return
	}
	fx.(func())()
}

func (s *Report) fxSource(t *T, dir string) (*lines.Events, *Testing) {
	return fxSourceDBG(t, s, dir)
}

func (s *Report) Passing_go_tests_only(t *T) {
	_, tt := s.fxSource(t, "go/pass")

	t.StarMatched(
		tt.afterWatch(awReporting).String(),
		fxExp["go/pass"]...,
	)
}

func TestReport(t *testing.T) {
	t.Parallel()
	Run(&Report{}, t)
}
