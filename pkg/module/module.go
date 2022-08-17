// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"os"
	"strings"
	"sync"
	"time"
)

// ErrNoModule is returned by [Module.Watch] in case set Module.Dir or
// the current working directory ascending to root doesn't contain a
// go.mod file.
var ErrNoModule = errors.New("module: no module found in path: ")

// DefaultInterval is the default value for Module.Interval which is
// used iff at the first call of [Module.Watch] no interval value is
// set.
var DefaultInterval = 200 * time.Millisecond

// DefaultIgnore is the default value for Module.Ignore a list of
// directories which is ignored when searching for testing packages in a
// go module's directory, e.g.:
//
//	m := Module{Ignore: append(DefaultIgnore, "my_additional_dir")}
//
// It is set iff Module.Ignore is unset at the first call of
// [Module.Watch].
var DefaultIgnore = []string{".git", "node_modules"}

// A Module represents a go module which can be watched for changes of
// testing packages.  A Module may not be copied after its first watcher
// has been registered by [Module.Watch].  A Module instance's methods
// may be used concurrently and arbitrary many watcher may be
// registered.
type Module struct {

	// mutex for a Module-instance's concurrency safety.
	mutex sync.Mutex

	// name of the go module represented by a Module-instance.
	name string

	// register channels messages to diff-reporting go routine
	// requesting to add send watcher.
	register chan *newWatcher

	// quit channels messages to diff-reporting go routing requesting
	// the removal of one or all watchers.  The former if send ID is
	// greater zero; the later if it is zero.
	quit chan uint64

	isWatched chan bool

	// newID creates a new Module-instance unique ID > 0 for registered
	// watcher to provide the possibility to remove a particular
	// registered watcher.  See Module.Quit
	newID func() uint64

	// The directory of a watched go module which is represented by a
	// Module instance.  If unset a Module.Watch initializes this
	// property by the first directory containing a go.mod file which is
	// found ascending along the path of the current working directory.
	Dir string

	// Interval is the duration between two packages diff-reports  for a
	// watcher.
	Interval time.Duration

	// Ignore is the list of directory names which are ignored in the
	// search for a go module's testing packages.  It defaults to
	// DefaultIgnore iff unset at the first call of Watch.  Note once
	// Module.Watch was called for the first time further modifications
	// of Ignore are not taken into account (until Module.QuitAll was
	// called and then Module.Watch again).
	Ignore []string
}

// Name returns a watched module's name.  Note if [Module.Watch] wasn't
// called the zero string is returned.
func (m *Module) Name() string {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.name
}

// Watch reports to each of its callers changes about a module's testing
// packages through returned channel.  The module which is reported
// about is the first module which is found in Module.Dir ascending
// towards root.  If Module.Dir is unset the current working directory
// is used.  If no directory with a go.mod file is found a wrapped
// ErrNoModule error is returned.  I.e. after this method's first call
// Module.Dir is the found module directory and [Module.Name] provides
// its name.  Returned ID may be used to unregister the watcher with
// given ID, see [Module.Quit].  If a watcher is unregistered its diff
// channel is closed.  See [Module.QuitAll] to learn how to release all
// resources acquired by this method.
func (m *Module) Watch() (diff <-chan *PackagesDiff, ID uint64, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := m.ensureNameAndDir(); err != nil {
		return nil, 0, err
	}
	if len(m.Ignore) == 0 {
		m.Ignore = DefaultIgnore
	}
	m.ensureDiffer() // go routine reporting diffs

	ID, _diff := m.newID(), make(chan *PackagesDiff, 1)
	m.register <- &newWatcher{
		diff: _diff,
		ID:   ID,
	}
	return _diff, ID, nil
}

func (m *Module) ensureNameAndDir() (err error) {
	if m.name != "" {
		return nil
	}
	dir := m.Dir
	if dir == "" {
		dir, err = os.Getwd()
		if err != nil {
			return err
		}
	}
	dir, name, err := findModule(dir)
	if err != nil {
		return err
	}
	m.name = name
	m.Dir = dir
	return nil
}

func (m *Module) ensureDiffer() {
	if m.register != nil {
		return
	}
	m.register = make(chan *newWatcher)
	m.quit = make(chan uint64)
	m.newID = idClosure()
	if m.Interval == 0 {
		m.Interval = DefaultInterval
	}
	m.isWatched = differ(m.Dir, m.Interval, ignoreClosure(m.Ignore...),
		m.register, m.quit)
}

// IsWatched returns true iff at least one watcher is registered.  Note
// a false return value doesn't mean that there is no diffing go routine
// running.  To guarantee this see [Module.QuitAll].
func (m *Module) IsWatched() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.isWatched == nil {
		return false
	}
	m.isWatched <- true
	return <-m.isWatched
}

// Quit unregisters the watcher with given ID and closes its
// diff-channel.  Quit is a no-op if no watcher with given ID exists.
func (m *Module) Quit(ID uint64) {
	if ID == 0 {
		return
	}
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.quit <- ID
}

// QuitAll closes all diff channels which were provided by
// [Module.Watch] and terminates the go routine reporting package diffs
// to watchers.
func (m *Module) QuitAll() {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	if m.quit == nil {
		return
	}
	close(m.quit)
	m.register = nil
	m.quit = nil
	m.isWatched = nil
}

func idClosure() func() uint64 {
	var ID uint64
	return func() uint64 {
		ID++
		return ID
	}
}

func ignoreClosure(ii ...string) func(string) bool {
	return func(s string) bool {
		for _, i := range ii {
			if !strings.HasSuffix(s, i) {
				continue
			}
			return true
		}
		return false
	}
}

type newWatcher struct {
	diff chan *PackagesDiff
	ID   uint64
}

type watcher struct {
	diff                     chan *PackagesDiff
	lastReported, lastPolled *packagesStat
}

// differ starts a go routine which every given interval informs all
// registered watchers about changes of testing packages in given
// directory (i.e. go module).  This go routine also listens on the
// register and quit channel to add a new watcher respectively remove
// one or all.  The later happens if the zero value is received over the
// quit channel.  NOTE the provided diff channel of a new watcher must
// be buffered with the capacity 1!  In this case the go routine can
// guarantee to not block and keep each watcher individually accurately
// posted about changes independently if a watcher is polling from its
// diff-channel or not.
func differ(
	dir string, interval time.Duration, ignore func(string) bool,
	register chan *newWatcher, quit <-chan uint64,
) (isWatched chan bool) {
	ww, isWatched := map[uint64]*watcher{}, make(chan bool)

	go func() {
		for {
			select {
			case <-isWatched:
				isWatched <- len(ww) > 0
			case register := <-register:
				ww[register.ID] = &watcher{diff: register.diff}
			case wID := <-quit:
				if terminate := quitWatching(wID, ww); terminate {
					return
				}
			case <-time.After(interval):
				reportDiffs(calcPackagesStat(dir, ignore), ww)
			}
		}
	}()

	return isWatched
}

func quitWatching(
	wID uint64, ww map[uint64]*watcher,
) (terminate bool) {

	if wID == 0 { // zero value means we quit all and terminate
		for _, w := range ww {
			close(w.diff)
		}
		return true
	}

	w, ok := ww[wID]
	if !ok {
		return
	}
	close(w.diff)
	delete(ww, wID)
	return
}

func reportDiffs(snapshot *packagesStat, ww map[uint64]*watcher) {
	for _, w := range ww {
		// w.diff has a 1-buffer which is drained ...
		select {
		case <-w.diff:
			// ... in case the value hasn't been read;
		default:
			// otherwise last reported becomes last polled...
			w.lastPolled = w.lastReported
		}

		diff := snapshot.diff(w.lastPolled)
		if diff == nil {
			continue
		}

		// ... and we send the most recent diff relative to the last
		// snapshot whose diff was polled to the watcher making sure at
		// the same time to not get blocked
		w.diff <- diff
		w.lastReported = snapshot
	}
}
