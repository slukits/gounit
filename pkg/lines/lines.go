// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package lines provides a simple terminal-UI where the UI is
// interpreted as an ordered set of lines.  Reported events are
//
// - rune: user pressed a keyboard key which translates to a rune
//
// - key: user pressed a keyboard key which doesn't translate to a
//   single rune
//
// - quit: user pressed q, ctrl-d or ctrl-c
//
// - resize: for initial layout or user resizes the terminal window.
//
// Event listeners are registered using a view's *Register* property.
// Calling the *Listen*-Method starts the event-loop with the mandatory
// resize-event and blocks until the quit event is received ('q',
// ctrl-c, ctrl-d).  A typical example:
//
//     lv := lines.NewView()
//     lv.Register.Resize(func(v *lines.View) {
//         v.For(func(l *lines.Line) {
//             l.Set(fmt.Sprintf("line %d", l.Idx+1))
//         })
//         lv.Get(0).Set(fmt.Sprintf("have %d lines", v.Len()))
//     })
//     // ... register other event-listeners for ...
//     lv.Listen() // block until quit
package lines

import (
	"errors"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/slukits/ints"
)

// View provides a line-based terminal user-interface.  A zero view is
// not ready to use.  *NewView* and *NewSim* create and initialize a new
// view whereas the later creates a view with tcell's simulation-screen
// for testing.  Typically you would set a new view's event-listeners
// followed by a (blocking) call of *Listen* starting the event-loop.
// Calling *Quit* stops the event-loop and releases resources.
type View struct {
	lib      tcell.Screen
	ll       []*Line
	evtRunes string
	evtKeys  ints.Set

	// Register holds a view's registered event listeners to which are
	// occurring events reported. (see type *Register*).
	Register ListenerRegister
}

// ListenerRegister allows to wrap a Register-instance and replace it
// with the wrapping which is useful for test-fixture creation (see
// testdata/fx.View)
type ListenerRegister interface {
	Rune(func(*View), ...rune) error
	Runes(func(*View, rune))
	Key(func(*View, tcell.ModMask), ...tcell.Key) error
	Resize(func(*View))
	Quit(func())
	reportRune(*View, rune)
	reportKey(*View, tcell.Key, tcell.ModMask)
	reportResize(*View)
	reportQuit(*tcell.EventKey) bool
}

// NewView returns a new View instance or nil and an error in case
// tcell's screen-creation or its initialization fails.
func NewView() (*View, error) {
	lib, err := screenFactory.NewScreen()
	if err != nil {
		return nil, err
	}
	if err := lib.Init(); err != nil {
		return nil, err
	}
	return &View{lib: lib, Register: newRegister()}, nil
}

// NewSim returns a new View instance wrapping tcell's simulation
// screen for testing purposes.  Since the wrapped tcell screen is
// private it is returned as well to facilitate desired mock-ups.
// NewSim fails iff tcell's screen-initialization fails.
func NewSim() (*View, tcell.SimulationScreen, error) {
	lib := screenFactory.NewSimulationScreen("")
	if err := lib.Init(); err != nil {
		return nil, nil, err
	}
	return &View{lib: lib, Register: newRegister()}, lib, nil
}

// Len returns the number of lines of a terminal screen.  Note len of
// the simulation screen defaults to 25.
func (v *View) Len() int {
	_, h := v.lib.Size()
	return h
}

// For calls ascending ordered for each line of registered view back.
func (v *View) For(cb func(*Line)) {
	for _, l := range v.ll {
		cb(l)
	}
}

// Listen starts the view's event-loop listening for user-input and
// blocks until a quit-event is received or *Quit* is called.
func (v *View) Listen() error {

	// BUG the following four lines needed to go into the resize event
	// v.lib.Clear()
	// if v.Listeners.Resize != nil {
	//     v.Listeners.Resize(v)
	// }

	for {

		// Poll event (blocking as I understand)
		ev := v.lib.PollEvent()

		// Process event
		switch ev := ev.(type) {
		case *tcell.EventResize:
			v.lib.Clear()
			v.ensureLines()
			v.Register.reportResize(v)
			v.lib.Sync()
		case *tcell.EventKey:
			if v.Register.reportQuit(ev) {
				v.Quit()
				return nil
			}
			if ev.Key() == tcell.KeyRune {
				v.Register.reportRune(v, ev.Rune())
				continue
			}
			v.Register.reportKey(v, ev.Key(), ev.Modifiers())
		default:
			return nil
		}
	}
}

// ensureLines adapts after a resize event the view's lines-slice.
func (v *View) ensureLines() {
	if len(v.ll) == v.Len() {
		return
	}
	if len(v.ll) > v.Len() {
		v.ll = v.ll[:v.Len()]
		return
	}
	for len(v.ll) < v.Len() {
		v.ll = append(v.ll, &Line{})
	}
}

// Quit the event-loop.
func (v *View) Quit() {
	v.lib.Fini()
}

// Register implements a View-instance's Register property storing the
// event-listeners for events provided by a View-instance.
type Register struct {
	keys     *sync.Mutex
	kk       map[tcell.Key]func(*View, tcell.ModMask)
	runes    *sync.Mutex
	rr       map[rune]func(*View)
	other    *sync.Mutex
	resize   func(*View)
	quit     func()
	allRunes func(*View, rune)
}

func newRegister() *Register {
	return &Register{
		keys:  &sync.Mutex{},
		kk:    map[tcell.Key]func(*View, tcell.ModMask){},
		runes: &sync.Mutex{},
		rr:    map[rune]func(*View){},
		other: &sync.Mutex{},
	}
}

// RegisterErr is returned by Register.Rune and Register.Key if a
// listener is registered for an already registered rune/key-event.
var RegisterErr = errors.New("event listener overwrites existing")

// Rune registers given listener for given runes failing iff one of the
// runes is already registered or is the quit-rune.  In the later case
// none of the runes is registered.  Setting rune listeners is
// concurrency save.  If the listener is nil given runes are
// unregistered.
func (rg *Register) Rune(listener func(*View), rr ...rune) error {
	rg.runes.Lock()
	defer rg.runes.Unlock()

	if listener == nil {
		for _, _r := range rr {
			delete(rg.rr, _r)
		}
		return nil
	}

	for _, _r := range rr {
		if _, ok := rg.rr[_r]; ok {
			return RegisterErr
		}
		if _r == 'q' {
			return RegisterErr
		}
	}

	for _, _r := range rr {
		rg.rr[_r] = listener
	}
	return nil
}

func (rg *Register) Runes(listener func(*View, rune)) {
	rg.other.Lock()
	defer rg.other.Unlock()
	rg.allRunes = listener
}

func (rg *Register) reportRune(v *View, r rune) {
	rg.other.Lock()
	if rg.allRunes != nil {
		rg.allRunes(v, r)
		rg.other.Unlock()
		return
	}
	rg.other.Unlock()
	rg.runes.Lock()
	defer rg.runes.Unlock()
	if _, ok := rg.rr[r]; !ok {
		return
	}
	rg.rr[r](v)
}

// Key registers given listener for given keys failing iff one of the
// keys is already registered or if they are the quit keys.  In the
// later case none of the keys is registered.  Setting key listeners is
// concurrency save.  If the listener is nil given keys are
// unregistered.
func (rg *Register) Key(
	listener func(*View, tcell.ModMask), kk ...tcell.Key,
) error {
	rg.keys.Lock()
	defer rg.keys.Unlock()

	if listener == nil {
		for _, k := range kk {
			delete(rg.kk, k)
		}
		return nil
	}

	for _, k := range kk {
		if _, ok := rg.kk[k]; ok {
			return RegisterErr
		}
		if k == tcell.KeyCtrlC || k == tcell.KeyCtrlD {
			return RegisterErr
		}
	}

	for _, k := range kk {
		rg.kk[k] = listener
	}
	return nil
}

func (rg *Register) reportKey(v *View, k tcell.Key, m tcell.ModMask) {
	rg.keys.Lock()
	defer rg.keys.Unlock()
	if _, ok := rg.kk[k]; !ok {
		return
	}
	rg.kk[k](v, m)
}

// Resize registers given listener for the resize event.
func (rg *Register) Resize(listener func(*View)) {
	rg.other.Lock()
	defer rg.other.Unlock()
	rg.resize = listener
}

func (rg *Register) reportResize(v *View) {
	rg.other.Lock()
	defer rg.other.Unlock()
	if rg.resize == nil {
		return
	}
	rg.resize(v)
}

// Quit registers given listener for the quit event which is triggered
// by 'r'-rune, ctrl-c and ctrl-d.
func (rg *Register) Quit(listener func()) {
	rg.other.Lock()
	defer rg.other.Unlock()
	rg.quit = listener
}

func (rg *Register) reportQuit(ev *tcell.EventKey) bool {
	rg.other.Lock()
	defer rg.other.Unlock()
	if ev.Key() == tcell.KeyRune && ev.Rune() != 'q' {
		return false
	}
	if ev.Key() != tcell.KeyRune && ev.Key() != tcell.KeyCtrlC &&
		ev.Key() != tcell.KeyCtrlD {
		return false
	}
	if rg.quit == nil {
		return true
	}
	rg.quit()
	return true
}

type defaultFactory struct{}

func (f *defaultFactory) NewScreen() (tcell.Screen, error) {
	return tcell.NewScreen()
}

func (f *defaultFactory) NewSimulationScreen(
	s string,
) tcell.SimulationScreen {
	return tcell.NewSimulationScreen(s)
}

type screenFactoryer interface {
	NewScreen() (tcell.Screen, error)
	NewSimulationScreen(string) tcell.SimulationScreen
}

var screenFactory screenFactoryer = &defaultFactory{}
