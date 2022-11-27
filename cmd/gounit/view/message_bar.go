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

func (mb *messageBar) OnInit(e *lines.Env) {
	mb.Dim().SetHeight(3)
	fmt.Fprint(e.LL(1), mb.dflt)
}

func (mb *messageBar) OnUpdate(e *lines.Env, data interface{}) {
	// type save because message bar update only allows string
	s, _ := data.(string)
	if s == "" {
		fmt.Fprint(e.LL(1), mb.dflt)
		return
	}
	fmt.Fprint(e.LL(1), s)
}
