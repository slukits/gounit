// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"github.com/gdamore/tcell/v2"
)

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

func (ll *lines) sync() {
	for _, l := range *ll {
		if !l.dirty {
			continue
		}
		l.sync()
	}
}

// Line represents a View's screen line.  A changed line content is
// automatically synchronized with the screen.  Note changes of a line
// are not concurrency save.
type Line struct {
	lib     tcell.Screen
	stale   string
	content string
	dirty   bool

	// Idx is the zero based line index corresponding with the screen
	// line.
	Idx int

	Style tcell.Style
}

// Set updates the content of a line.
func (l *Line) Set(content string) *Line {
	if content == l.content {
		return l
	}
	if !l.dirty {
		l.dirty = true
	}
	if l.stale == "" {
		l.stale = l.content
	}
	l.content = content
	return l
}

func (l *Line) sync() {
	l.dirty = false
	if len(l.content) >= len(l.stale) {
		l.setLonger()
	} else {
		l.setShorter()
	}
	l.stale = ""
}

func (l *Line) setShorter() {
	base, add := len(l.content), len(l.stale)-len(l.content)
	l.setLonger()
	for i := 0; i < add; i++ {
		l.lib.SetContent(base+i, l.Idx, ' ', nil, l.Style)
	}
}

func (l *Line) setLonger() {
	for i, r := range l.content {
		l.lib.SetContent(i, l.Idx, r, nil, l.Style)
	}
}

// IsDirty returns true if a line content has changed since the last
// screen synchronization.
func (l *Line) IsDirty() bool {
	return l.dirty
}
