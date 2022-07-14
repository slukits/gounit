package lines

import "github.com/gdamore/tcell/v2"

// EventMask types flags classifying keys/runes for "internal" event
// handling.
type EventMask uint64

const (
	// QuiteEvent classifies keys/runes registered for the quite event.
	QuitEvent EventMask = 1 << iota
	UpEvent
	DownEvent
	// NoEvent classifies keys/runes not registered for any "internal"
	// event.
	NoEvent EventMask = 0
)

var InternalEvents = []EventMask{QuitEvent, UpEvent, DownEvent}

// Keys provides information about keys/runes which are registered for
// "internal" events.  It also allows to change these in a consistent
// way.  A *Register* instance is always with a copy of the
// *DefaultKeys* Keys-instance initialized.
type Keys struct {
	keys  map[tcell.Key]EventMask
	runes map[rune]EventMask
}

// DefaultKeys are the default runes and keys which are used for
// internal event handling
var DefaultKeys = &Keys{
	keys: map[tcell.Key]EventMask{
		tcell.KeyCtrlC: QuitEvent,
		tcell.KeyCtrlD: QuitEvent,
		tcell.KeyUp:    UpEvent,
		tcell.KeyDown:  DownEvent,
	},
	runes: map[rune]EventMask{
		'q': QuitEvent,
		'k': UpEvent,
		'j': DownEvent,
	},
}

func (kk *Keys) copy(reg *Register) *Keys {
	cpy := Keys{
		keys:  make(map[tcell.Key]EventMask),
		runes: map[rune]EventMask{},
	}
	for k, v := range kk.keys {
		cpy.keys[k] = v
	}
	for k, v := range kk.runes {
		cpy.runes[k] = v
	}
	return &cpy
}

// KeysOf returns the keys for given "internally" handled event.
func (kk *Keys) KeysOf(e EventMask) []tcell.Key {
	_kk := []tcell.Key{}
	for k, _e := range kk.keys {
		if e != _e {
			continue
		}
		_kk = append(_kk, k)
	}
	return _kk
}

// RunesOf returns the runes for given "internally" handled event.
func (kk *Keys) RunesOf(e EventMask) []rune {
	rr := []rune{}
	for r, _e := range kk.runes {
		if e != _e {
			continue
		}
		rr = append(rr, r)
	}
	return rr
}

// HasKey returns true if given key is registered for internal event
// handling.
func (kk *Keys) HasKey(k tcell.Key) bool {
	return kk.keys[k] != NoEvent
}

// HasRune returns true if given rune is registered for internal event
// handling.
func (kk *Keys) HasRune(r rune) bool {
	return kk.runes[r] != NoEvent
}

// KeyEvent maps a key to its internally handled event or to NoEvent if
// not registered.
func (kk *Keys) KeyEvent(k tcell.Key) EventMask {
	return kk.keys[k]
}

// RuneEvent maps a rune to its internally handled event or to NoEvent
// if not registered.
func (kk *Keys) RuneEvent(r rune) EventMask {
	return kk.runes[r]
}
