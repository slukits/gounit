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
	dflt     string
	listener func(int, LLMod)
}

func (m *main) OnInit(e *lines.Env) {
	fmt.Fprint(e, m.dflt)
}

func (m *main) OnClick(_ *lines.Env, _, y int) {
	if m.listener == nil {
		return
	}
	m.listener(y, Default)
}

func (m *main) OnContext(_ *lines.Env, _, y int) {
	if m.listener == nil {
		return
	}
	m.listener(y, Context)
}

func (m *main) OnUpdate(e *lines.Env) {
	data := e.Evt.(*lines.UpdateEvent).Data
	switch dt := data.(type) {
	case map[int]string:
		clear := false
		for idx, content := range dt {
			if idx == -1 {
				clear = true
				continue
			}
			fmt.Fprint(e.LL(idx), content)
		}
		if clear {
			for i := 0; i < m.Len(); i++ {
				if _, ok := dt[i]; ok {
					continue
				}
				m.Reset(i)
			}
		}
	case func(int, LLMod):
		m.listener = dt
	}
}
