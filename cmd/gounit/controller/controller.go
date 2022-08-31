// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/lines"
)

const (
	WatcherErr = "gounit: watcher: %s: %v"
)

// Events is a function to initialize the terminal ui with components
// returning an Events-instance to listen for events.
type Events func(lines.Componenter) *lines.Events

// Watcher is a function whose returned channel watches a go modules
// packages sources whose tests runs are reported to a terminal ui.
type Watcher interface {
	ModuleName() string
	ModuleDir() string
	SourcesDir() string
	Watch() (<-chan *model.PackagesDiff, uint64, error)
}

// New starts the application and blocks until a quit event occurs.
// Fatale errors are reported to ftl while ll is used to initialize the
// ui and start the event loop.
func New(ftl func(...interface{}), w Watcher, ee Events) {
	diff, _, err := w.Watch()
	if err != nil {
		ftl(WatcherErr, w.SourcesDir(), err)
		return
	}
	_ = diff
	ee(nil).Listen()
}
