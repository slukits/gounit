// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"github.com/gdamore/tcell/v2"
)

type ErrScr struct {
	lib     tcell.Screen
	content string
	isDirty bool
	Active  bool
	Style   tcell.Style
}

func (v *Screen) ErrScreen() *ErrScr {
	if v.errScr == nil {
		v.errScr = &ErrScr{lib: v.lib}
	}
	return v.errScr
}

func (e *ErrScr) IsDirty() bool {
	return e.isDirty
}

func (e *ErrScr) String() string {
	return e.content
}

func (e *ErrScr) Set(s string) {
	if s == e.content {
		return
	}
	if !e.isDirty {
		e.isDirty = true
	}
	e.content = s
}

func (e *ErrScr) sync() {
	e.isDirty = false
	e.lib.Clear()
	w, h := e.lib.Size()
	y := h / 2
	x := w/2 - len(e.content)/2
	if len(e.content) > w {
		x = 0
	}
	for i, r := range e.content {
		e.lib.SetContent(x+i, y, r, nil, e.Style)
	}
}
