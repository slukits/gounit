// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"

	"github.com/slukits/lines"
)

type main struct {
	lines.Component
	dflt string
}

func (m *main) OnInit(env *lines.Env) {
	fmt.Fprint(env, m.dflt)
}
