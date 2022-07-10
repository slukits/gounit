// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"testing"
	"time"
)

type timeStepper struct{ Suite }

func (s *timeStepper) SetUp(t *T) { t.Parallel() }

func (s *timeStepper) Duration_defaults_to_10_milliseconds(t *T) {
	var ts TimeStepper
	t.Eq(10*time.Millisecond, ts.Duration())
}

func (s *timeStepper) Step_defaults_to_1_millisecond(t *T) {
	var ts TimeStepper
	t.Eq(1*time.Millisecond, ts.Step())
}

func (s *timeStepper) Sets_duration(t *T) {
	var ts TimeStepper
	t.Eq(5*time.Millisecond,
		ts.SetDuration(5*time.Millisecond).Duration())
}

func (s *timeStepper) Sets_step(t *T) {
	var ts TimeStepper
	t.Eq(5*time.Millisecond, ts.SetStep(5*time.Millisecond).Step())
}

func (s *timeStepper) Step_adding_returns_false_if_no_more_steps(t *T) {
	ts := (&TimeStepper{}).SetDuration(1 * time.Millisecond)
	t.False(ts.AddStep())
}

func TestTimeStepper(t *testing.T) {
	t.Parallel()
	Run(&timeStepper{}, t)
}
