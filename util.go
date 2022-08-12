// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"sync"
	"time"
)

// Fixtures provides a simple concurrency save fixture storage for
// gounit tests.  A Fixtures instance must not be copied after its first
// use.  A Fixtures storage is typically used to setup test specific
// fixtures for concurrently run suite-tests
//
//	type MySuite {
//	    gounit.Suite
//	    fx ff
//	}
//
//	type ff { gounit.Fixtures }
//
//	func (fx *ff) Of(t *gounit.T) string { return fx.Get(t).(string) }
//
//	func (s *MySuite) SetUp(t *gounit.T) {
//	    t.Parallel()
//	    s.fx.Set(t, fmt.Sprintf("%p's fixture", t))
//	}
//
//	func (s *MySuite) MySuiteTest(t *gounit.T) {
//	    t.Logf("%p: got: %s", t, s.fx.Of(t))
//	}
//
//	func TestMySuite(t *testing.T) {
//	    t.Parallel()
//	    Run(&MySuite{}, t)
//	}
type Fixtures struct {
	mutex sync.Mutex
	ff    map[*T]interface{}
}

// Set adds concurrency save a mapping from given test to given fixture.
func (ff *Fixtures) Set(t *T, fixture interface{}) {
	ff.mutex.Lock()
	defer ff.mutex.Unlock()
	if ff.ff == nil {
		ff.ff = map[*T]interface{}{}
	}
	ff.ff[t] = fixture
}

// Get maps given test to its fixture and returns it.
func (ff *Fixtures) Get(t *T) interface{} {
	ff.mutex.Lock()
	defer ff.mutex.Unlock()
	return ff.ff[t]
}

// Del removes the mapping of given test to its fixture and returns the
// fixture.
func (ff *Fixtures) Del(t *T) interface{} {
	ff.mutex.Lock()
	defer ff.mutex.Unlock()
	fixture := ff.ff[t]
	delete(ff.ff, t)
	return fixture
}

// TimeStepper provides the features to split a duration into segments.
// The duration defaults to 10 milliseconds segmented into 1 millisecond
// steps.  The zero value is ready to use.
type TimeStepper struct {
	duration time.Duration
	step     time.Duration
	elapsed  time.Duration
}

// Duration is the overall duration a time-stepper represents defaulting
// to 10 milliseconds.
func (t *TimeStepper) Duration() time.Duration {
	if t.duration == 0 {
		t.duration = 10 * time.Millisecond
	}
	return t.duration
}

// SetDuration sets the overall duration a time-stepper represents.
func (t *TimeStepper) SetDuration(d time.Duration) *TimeStepper {
	t.duration = d
	return t
}

// Step is the step-segment of a time-stepper's overall duration
// defaulting to 1 millisecond.
func (t *TimeStepper) Step() time.Duration {
	if t.step == 0 {
		t.step = 1 * time.Millisecond
	}
	return t.step
}

// SetStep sets the duration of a segment of a time-stepper's the
// overall duration.
func (t *TimeStepper) SetStep(s time.Duration) *TimeStepper {
	t.step = s
	return t
}

// AddStep adds an other step to the elapsed time and returns true if
// there is still time left; false otherwise.
func (t *TimeStepper) AddStep() bool {
	t.elapsed += t.Step()
	return t.Duration() > t.elapsed
}
