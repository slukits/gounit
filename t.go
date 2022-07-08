// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"fmt"
	"testing"
)

// T wraps a *testing.T*-instance into a gounit.T instance which adjusts
// testing.T's api.
type T struct {
	Idx    int
	t      *testing.T
	logger func(...interface{})
}

// Log writes given arguments to set logger which defaults to the logger
// of wrapped *testing.T* instance.  The default may be overwritten by a
// suite-embedder implementing the SuiteLogging interface.
func (t *T) Log(args ...interface{}) { t.logger(args...) }

// Logf writes given format string leveraging Sprintf to set logger which
// defaults to the logger of wrapped *testing.T* instance.  The default
// may be overwritten by a suite-embedder implementing the SuiteLogging
// interface.
func (t *T) Logf(format string, args ...interface{}) {
	t.Log(fmt.Sprintf(format, args...))
}
