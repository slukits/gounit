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

// Line represents a View's screen line.  A changed line content is
// automatically synchronized with the screen.  Note changes of a line
// are not concurrency save.
type Line struct {
	lib     tcell.Screen
	content string
	dirty   bool

	// Idx is the zero based line index corresponding with the screen
	// line.
	Idx int
}

// Set updates the content of a line.
func (l *Line) Set(content string) *Line {
	if content == l.content {
		return l
	}
	if !l.dirty {
		l.dirty = true
	}

	if len(content) >= len(l.content) {
		return l.setLonger(content)
	}
	return l.setShorter(content)
}

func (l *Line) setShorter(content string) *Line {
	base, add := len(content), len(l.content)-len(content)
	l.setLonger(content)
	for i := 0; i < add; i++ {
		l.lib.SetContent(base+i, l.Idx, ' ', nil, tcell.StyleDefault)
	}
	return l
}

func (l *Line) setLonger(content string) *Line {
	l.content = content
	for i, r := range content {
		l.lib.SetContent(i, l.Idx, r, nil, tcell.StyleDefault)
	}
	return l
}

// IsDirty returns true if a line content has changed since the last
// screen synchronization.
func (l *Line) IsDirty() bool {
	return l.dirty
}
