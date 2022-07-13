// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"github.com/gdamore/tcell/v2"
)

// Zero is the zero line which has no functionality other than being
// provided if a line with an out of bound index is requested.
var Zero = &Line{}

// Line represents a View's screen line.  A changed line content is
// automatically synchronized with the screen.  Note changes of a line
// are not concurrency save.
type Line struct {
	lib     tcell.Screen
	stale   string
	content string
	dirty   bool
	typ     int

	// Idx is the zero based line index corresponding with the screen
	// line.
	Idx int

	Style tcell.Style
}

// Set updates the content of a line.
func (l *Line) Set(content string) *Line {
	if content == l.content || l.typ == 0 {
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

// Get returns the currently set line content while stale if not zero is
// the lines content on the screen.
func (l *Line) Get() (current, stale string) {
	return l.content, l.stale
}

// SetType sets a line's type which must be bigger than the
// *DefaultType*.
func (l *Line) SetType(t int) (ok bool) {
	if t < DefaultType || l.typ == 0 {
		return false
	}
	l.typ = t
	return true
}

// Type returns a lines type.
func (l *Line) Type() int {
	return l.typ
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
	if l.typ == 0 {
		return false
	}
	return l.dirty
}
