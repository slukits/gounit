// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

type Lines struct {
	scr      *Screen
	ll       []*Line
	scrFirst int
}

// Len returns the number of lines of lines-view.  It defaults to a
// views Len which is its minimum.
func (ll *Lines) Len() int { return len(ll.ll) }

// For calls in ascending order for each line of its lines back.
func (ll *Lines) For(cb func(*Line)) {
	for _, l := range ll.ll {
		cb(l)
	}
}

// FirstScreenLine returns the index of the first of the currently
// focused Lines-component.
func (ll *Lines) FirstScreenLine() int { return ll.scrFirst }

// SetFirstScreenLine sets first line shown on the screen of the
// currently focused Lines-component.
func (ll *Lines) SetFirstScreenLine(i int) *Lines {
	if i < 0 || i >= len(ll.ll) {
		return ll
	}
	ll.scrFirst = i
	return ll
}

// ForScreen calls ascending ordered back for each line shown on the
// screen.
func (ll *Lines) ForScreen(cb func(*Line)) {
	bound := ll.scr.Len()
	if len(ll.ll)-ll.scrFirst < bound {
		bound = len(ll.ll) - ll.scrFirst
	}
	for i := ll.scrFirst; i < bound; i++ {
		cb(ll.ll[i])
	}
}

// ForN calls back for n lines creating needed lines if n > Len.
func (ll *Lines) ForN(n int, cb func(*Line)) {
	if n <= 0 {
		return
	}
	ll.ensure(n)
	for i := 0; i < n; i++ {
		cb(ll.ll[i])
	}
}

// Line returns line with given index or the zero line if no such line
// exists.
func (ll *Lines) Line(idx int) *Line {
	if idx < 0 || idx >= len(ll.ll) {
		return Zero
	}
	return ll.ll[idx]
}

const DefaultType = 1

// ensure makes sure that there are at least n lines.
func (ll *Lines) ensure(n int) {
	lower, m := len(ll.ll), 0
	for len(ll.ll) < n {
		ll.ll = append(ll.ll, &Line{
			lib: ll.scr.lib,
			typ: DefaultType,
			Idx: lower + m})
		m++
	}
}

func (ll *Lines) isDirty() bool {
	for _, l := range ll.ll {
		if !l.dirty {
			continue
		}
		return true
	}
	return false
}

func (ll *Lines) sync() {
	for _, l := range ll.ll {
		if !l.dirty {
			continue
		}
		l.sync()
	}
}
