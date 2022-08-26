// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"

	"github.com/slukits/lines"
)

type messageBar struct {
	lines.Component
	dflt string
}

func (mb *messageBar) OnInit(env *lines.Env) {
	mb.Dim().SetHeight(3)
	fmt.Fprint(env, mb.dflt)
}

func (mb *messageBar) OnUpdate(e *lines.Env) {
	// type save because message bar update only allows string
	s, _ := e.Evt.(*lines.UpdateEvent).Data.(string)
	if s == "" {
		fmt.Fprint(e, mb.dflt)
		return
	}
	fmt.Fprint(e, s)
}
