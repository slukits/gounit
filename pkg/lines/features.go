// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package lines

import (
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/ints"
)

// Feature classifies keys/runes for "internal" event handling.
type Feature uint64

const (
	// NoFeature classifies keys/runes not registered for any "internal"
	// event.
	NoFeature Feature = 0
	// FtQuit classifies keys/runes registered for the quite event.
	FtQuit Feature = iota
	// FtUp classifies keys/runes registered for scrolling up.
	FtUp
	// FtDown classifies keys/runes registered for scrolling down.
	FtDown
)

// InternalFeatures provides a slice of all the potentially internally
// handled features
var InternalFeatures = []Feature{FtQuit, FtUp, FtDown}

// Features provides information about keys/runes which are registered
// for features provided by the lines-package.  It also allows to change
// these in a consistent and convenient way.  The zero value is not
// ready to use.  Make a copy of DefaultFeatures to create a new
// Features-instance.  Note  A *Register* instance is
// always with a copy of the *DefaultFeatures* Features-instance
// initialized which holds the quit-feature only.
type Features struct {
	mutex *sync.Mutex
	keys  map[tcell.ModMask]map[tcell.Key]Feature
	runes map[rune]Feature
}

// DefaultFeatures are the default runes and keys which are associated
// with internally handled events.  NOTE DefaultFeatures cannot be
// modified, a copy of them can!
var DefaultFeatures = &Features{
	keys: map[tcell.ModMask]map[tcell.Key]Feature{
		tcell.ModNone: {
			tcell.KeyCtrlC: FtQuit,
			tcell.KeyCtrlD: FtQuit,
		},
	},
	runes: map[rune]Feature{
		0:   NoFeature,
		'q': FtQuit,
	},
}

// modifiable returns false for the default features.
func (ff *Features) modifiable() bool {
	_, ok := ff.runes[0]
	return ok
}

// Copy creates a new Features instance initialized with the features of
// receiving Features instance.
func (ff *Features) Copy() *Features {
	cpy := Features{
		keys:  make(map[tcell.ModMask]map[tcell.Key]Feature),
		runes: map[rune]Feature{},
	}
	for m, kk := range ff.keys {
		cpy.keys[m] = map[tcell.Key]Feature{}
		for k, f := range kk {
			cpy.keys[m][k] = f
		}
	}
	for r, f := range ff.runes {
		cpy.runes[r] = f
	}
	return &cpy
}

// Add associates given feature with given rune, key and modifier.
// Provide the respective zero-values for arguments you don't want to
// provide.  ModMasks are only associated with keys.
func (ff *Features) Add(f Feature, r rune, k tcell.Key, m tcell.ModMask) {
	if !ff.modifiable() {
		return
	}
}

// Del removes all keys and runes registered for given feature except
// for the quit feature.  In the later case only registered runes are
// removed.  ctrl-c and ctrl-d are always registered for the quit key.
func (ff *Features) Del(f Feature) {
	if !ff.modifiable() {
		return
	}
}

// Registered returns the set of features currently registered.
func (ff *Features) Registered() *ints.Set {
	_ff := &ints.Set{}
	for _, kk := range ff.keys {
		for _, f := range kk {
			_ff.Add(int(f))
		}
	}
	for _, f := range ff.runes {
		if f == NoFeature {
			continue
		}
		_ff.Add(int(f))
	}
	return _ff
}

type FeatureKey struct {
	Mod tcell.ModMask
	Key tcell.Key
}

// KeysOf returns the keys with their modifiers for given feature.
func (ff *Features) KeysOf(f Feature) []*FeatureKey {
	kk := []*FeatureKey{}
	for m, _kk := range ff.keys {
		for k, _f := range _kk {
			if f != _f {
				continue
			}
			kk = append(kk, &FeatureKey{Mod: m, Key: k})
		}
	}
	return kk
}

// RunesOf returns the runes for given lines-feature.
func (kk *Features) RunesOf(e Feature) []rune {
	rr := []rune{}
	for r, _e := range kk.runes {
		if e != _e {
			continue
		}
		rr = append(rr, r)
	}
	return rr
}

// HasKey returns true if given key is registered for a feature.
func (kk *Features) HasKey(k tcell.Key, m tcell.ModMask) bool {
	return kk.keys[m][k] != NoFeature
}

// HasRune returns true if given rune is registered for internal event
// handling.
func (kk *Features) HasRune(r rune) bool {
	return kk.runes[r] != NoFeature
}

// KeyEvent maps a key to its internally handled event or to NoEvent if
// not registered.
func (kk *Features) KeyEvent(k tcell.Key, m tcell.ModMask) Feature {
	return kk.keys[m][k]
}

// RuneEvent maps a rune to its internally handled event or to NoEvent
// if not registered.
func (kk *Features) RuneEvent(r rune) Feature {
	return kk.runes[r]
}
