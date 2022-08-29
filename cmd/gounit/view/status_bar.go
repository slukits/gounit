// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"

	"github.com/slukits/lines"
)

type statusBar struct {
	lines.Component
	dflt string
}

func (sb *statusBar) OnInit(e *lines.Env) {
	sb.Dim().SetHeight(2)
	if sb.dflt != "" {
		fmt.Fprint(e.LL(1), sb.dflt)
	}
}

func (mb *statusBar) OnUpdate(e *lines.Env) {
	// type save because message bar update only allows string
	s, _ := e.Evt.(*lines.UpdateEvent).Data.(string)
	if s == "" {
		fmt.Fprint(e.LL(1), mb.dflt)
		return
	}
	fmt.Fprint(e.LL(1), s)
}
