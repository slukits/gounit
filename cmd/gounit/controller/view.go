// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
view.go contains the functionality of the controller needed to receive
user requests and update the view in response to a user request or a
model update.  I.e. there are two sources of change for the view: user
input and model change.  Hence the view is updated through a single
(locking) function.
*/

package controller

import (
	"fmt"
	"strings"
	"sync"

	"github.com/slukits/gounit/cmd/gounit/view"
)

const initReport = "waiting for testing packages being reported..."

// viewIniter instance implements view.Initer, i.e. provides the initial
// data to a new view and collects the provided view modifiers.
type viewIniter struct {

	// controller instance providing needed information to initialize a
	// view.
	controller *controller

	// ftl to report fatal errors during the initialization process.
	ftl func(...interface{})
}

func (i *viewIniter) Fatal() func(...interface{}) { return i.ftl }

func (i *viewIniter) Message(msg func(string)) string {
	i.controller.view.msg = msg
	return fmt.Sprintf("%s: %s",
		i.controller.watcher.ModuleName(), strings.TrimPrefix(
			i.controller.watcher.SourcesDir(),
			i.controller.watcher.ModuleDir(),
		),
	)
}

func (i *viewIniter) Reporting(upd func(view.Reporter)) view.Reporter {
	i.controller.view.rprUpd = upd
	return &report{ll: []string{initReport}}
}

func (i *viewIniter) Status(upd func(view.Statuser)) {
	i.controller.view.sttUpd = upd
}

func (i *viewIniter) Buttons(upd func(view.Buttoner)) view.Buttoner {
	i.controller.view.bttUpd = upd
	return i.controller.bb.defaults()
}

// viewUpdater collects the functions to update aspects of a view which are
// used by the viewUpdater to modify the view.
type viewUpdater struct {

	// Mutex avoids that the view is updated concurrently.
	*sync.Mutex

	// msg updates the view's message bar
	msg func(string)

	// sttUpd updates the view's status bar
	sttUpd func(view.Statuser)

	// rprUpd updates lines of a view's reporting component.
	rprUpd func(view.Reporter)

	// bttUpd updates the buttons of a view's button bar.
	bttUpd func(view.Buttoner)
}

// Update updates the view and should be the only way the view is
// updated to avoid data races.
func (vw *viewUpdater) Update(dd ...interface{}) {
	vw.Lock()
	defer vw.Unlock()

	for _, d := range dd {
		switch updData := d.(type) {
		case view.Buttoner:
			vw.bttUpd(updData)
		case view.Reporter:
			vw.rprUpd(updData)
		case *view.Statuser:
			vw.sttUpd(*updData)
		}
	}
}
