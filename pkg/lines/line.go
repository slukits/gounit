// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import "github.com/gdamore/tcell/v2"

type lines []*Line

func (ll *lines) isDirty() bool {
	for _, l := range *ll {
		if !l.dirty {
			continue
		}
		return true
	}
	return false
}

func (ll *lines) clean() {
	for _, l := range *ll {
		if !l.dirty {
			continue
		}
		l.dirty = false
	}
}

type Line struct {
	lib     tcell.Screen
	content string
	dirty   bool
	Idx     int
}

func (l *Line) Set(content string) *Line {
	if content == l.content {
		return l
	}
	if !l.dirty {
		l.dirty = true
	}
	l.content = content
	for i, r := range content {
		l.lib.SetContent(i, l.Idx, r, nil, tcell.StyleDefault)
	}
	return l
}

func (l *Line) IsDirty() bool {
	return l.dirty
}
