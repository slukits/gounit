// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"errors"
	"os"
	"sync"
	"time"
)

var ErrNoModule = errors.New("module: no module found in path: ")

var DefaultInterval = 200 * time.Millisecond

type Event uint

const (
	FoundModule Event = iota
)

func idClosure() func() uint64 {
	var ID uint64
	return func() uint64 {
		ID++
		return ID
	}
}

// A Module represents a go module which can be watched for changes of
// testing packages.  A Module may not be copied after its first watcher
// has been registered by [Module.Watch].  A Module instance may be used
// concurrently and arbitrary many watcher may be registered.
type Module struct {

	// mutex makes module concurrency save
	mutex sync.Mutex

	// name is the module name.
	name string

	// register channels messages to diff-reporting go routine
	// requesting to add given watcher.
	register chan *newWatcher

	// quit channels messages to diff-reporting go routing requesting
	// the removal of one or all watchers.
	quit chan uint64

	// newID creates a new Module-instance unique ID > 0 for registered
	// watcher to provide the possibility to remove a particular
	// registered watcher
	newID func() uint64

	// Dir the module's directory.  If unset a Watch initializes this
	// property by the first directory containing a go.mod file which is
	// found along the path of the current working directory.
	Dir string

	// Interval is the duration after which a packages snapshot is
	// compared with the last packages snapshot.  If not set at first
	// call of Watch it defaults to [module.DefaultInterval].
	Interval time.Duration
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
// ErrNoModule error is returned.  I.e. after the first call of Watch
// Dir is the found module directory and [Module.Name] provides its
// name.  Returned ID may be used to unregister the watcher with given
// ID, see [Module.Quit].  If a watcher is unregistered its diff channel
// is closed.  See [Module.QuitAll] to learn how to release all resources.
func (m *Module) Watch() (diff <-chan *PackagesDiff, ID uint64, err error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if err := m.ensureNameAndDir(); err != nil {
		return nil, 0, err
	}
	m.ensureDiffer()

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
	go differ(m.Dir, m.Interval, m.register, m.quit)
}

// IsWatched returns true if at least one watcher is registered;
// otherwise false.
func (m *Module) IsWatched() bool {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.register != nil
}

// Quit unregisters the watcher with given ID.  Quit is a no-op if no
// watcher with given ID exists.
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
}

type newWatcher struct {
	diff chan *PackagesDiff
	ID   uint64
}

type watcher struct {
	diff                        chan *PackagesDiff
	lastReported, lastRetrieved packages
}

// differ starts a go routine which every given interval informs all
// registered watchers about changes of testing packages in given
// directory.  This go routine also listens on the the register channel
// and quit channel to add a new watcher respectively remove one or all.
// The later happens if the zero value is received over the quit
// channel.  NOTE the provided diff channel of a new watcher must be
// buffered with the capacity 1!  In this case the go routine can
// guarantee to not block and keep each watcher individually accurately
// posted about changes independent of a watcher reading a send diff or
// not.
func differ(
	dir string, interval time.Duration,
	register chan *newWatcher, quit <-chan uint64,
) {
	ww := map[uint64]*watcher{}

	go func() {
		for {
			select {
			case register := <-register:
				ww[register.ID] = &watcher{diff: register.diff}
			case wID := <-quit:
				if wID == 0 { // zero value means we quit all and return
					for _, w := range ww {
						close(w.diff)
					}
					return
				} // else a particular watcher should be removed
				w, ok := ww[wID]
				if !ok {
					continue
				}
				close(w.diff)
				delete(ww, wID)
			case <-time.After(interval):
				snapshot := packagesSnapshot(dir)
				for _, w := range ww {
					// w.diff has a 1-buffer which is drained ...
					select {
					case <-w.diff:
						// ... in case the value hasn't been read;
					default:
						// otherwise last reported becomes last
						// retrieved...
						w.lastRetrieved = w.lastReported
					}

					diff := snapshot.diff(w.lastRetrieved)
					if diff == nil {
						continue
					}

					// ... and send the most recent diff relative to the
					// last snapshot whose diff was retrieved to the
					// watcher making sure at the same time to not get
					// blocked
					w.diff <- diff
					w.lastReported = snapshot
				}
			}
		}
	}()
}
