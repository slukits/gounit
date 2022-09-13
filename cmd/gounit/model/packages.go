// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// PackagesDiff reports the differences of a module's packages at two
// different points in time.  PackagesDiff is immutable hence we can
// report an instance by sending a pointer over an according channel.
type PackagesDiff struct {
	last, current *packagesStat
	timeout       time.Duration
}

// For returns all testing packages which were updated since the last
// reported diff in descending order by their modification time.
func (d *PackagesDiff) For(cb func(*TestingPackage) (stop bool)) error {
	sort.Slice(d.current.pp, func(i, j int) bool {
		return d.current.pp[i].ModTime.After(d.current.pp[j].ModTime)
	})
	for _, ps := range d.current.pp {
		if !d.last.isUpdatedBy(ps) {
			continue
		}
		tt, err := ps.loadTestFiles()
		if err != nil {
			return err
		}
		tp := &TestingPackage{
			ModTime: ps.ModTime,
			id:      ps.rel, abs: ps.abs, files: tt, timeout: d.timeout}
		if cb(tp) {
			return nil
		}
	}
	return nil
}

// ForDel returns a testing package which got deleted.  Note neither
// tests nor suites are provide by such a testing package.
func (d *PackagesDiff) ForDel(cb func(*TestingPackage) (stop bool)) {
	if d.last == nil || len(d.last.pp) == 0 {
		return
	}
	for _, ps := range d.last.pp {
		if d.current.has(ps) {
			continue
		}
		tp := &TestingPackage{
			id: ps.rel, abs: ps.abs, parsed: true, timeout: 0}
		if cb(tp) {
			return
		}
	}
}

// hasDelta returns true iff the two packages stats represent different
// numbers of package stats or if a current package stat updates (or
// misses) its corresponding last package stat.
func (d *PackagesDiff) hasDelta() bool {
	if d.last == nil && d.current == nil {
		return false
	}
	if d.last == nil || d.current == nil {
		return true
	}
	if len(d.last.pp) != len(d.current.pp) {
		return true
	}
	for _, p := range d.current.pp {
		if !d.last.isUpdatedBy(p) {
			continue
		}
		return true
	}
	// NOTE the case that d.last contains package stats which are not in
	// d.current can be safely neglected since at this point d.last and
	// d.current have equally many package stats, i.e. if d.current
	// misses package stats of d.last it must have package stats which
	// are not in d.last which trigger a true return value in the above
	// case already.
	return false
}

// A packagesStat holds a modules testing packages stats and the
// modification time of the most recently modified testing package.  It
// is used by [Module.PackagesDiff] to report changes between two
// packages stats of the same module at different points in time.
type packagesStat struct {
	ModTime time.Time
	pp      []*pkgStat
}

// diff returns a new PackagesDiff iff there are differences between
// receiving packages stats and given other packages stats.  I.e. a
// PackagesDiff instance is returned both packagesStat instances
// represent different series of package stats or if other contains a
// package stats with more recent modification time than its
// corresponding package stats in receiving packages stats.
func (pp *packagesStat) diff(other *packagesStat) *PackagesDiff {
	if pp == other {
		return nil
	}
	d := &PackagesDiff{last: other, current: pp}
	if !d.hasDelta() {
		return nil
	}
	return d
}

// isUpdatedBy returns true iff given package stats not included in
// receiving packages stats or if its modification time is after the
// modification time of corresponding package stats of receiving
// packages stats.
func (pp *packagesStat) isUpdatedBy(tp *pkgStat) bool {
	if pp == nil {
		return true
	}
	for _, _tp := range pp.pp {
		if _tp.ID() != tp.ID() {
			continue
		}
		if !tp.ModTime.After(_tp.ModTime) {
			return false
		}
	}
	return true
}

// has returns true iff receiving packages stats have package stats with
// the same relative name as given package stats.
func (pp *packagesStat) has(tp *pkgStat) bool {
	for _, _tp := range pp.pp {
		if _tp.ID() != tp.ID() {
			continue
		}
		return true
	}
	return false
}

// dirStack is a lifo stack of directory names to traverse a go module's
// directory structure without recursion.
type dirStack []string

// Push adds given directory to the stack.
func (s *dirStack) Push(d string) { *s = append(*s, d) }

// Pop returns and removes the last entry from given stack.
func (s *dirStack) Pop() string {
	l := len(*s)
	str := (*s)[l-1]
	*s = (*s)[:l-1]
	return str
}

// PushDir pushes each directory from given directory onto given stack.
func (s *dirStack) PushDir(dir string, ignore func(string) bool) error {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if !e.IsDir() || ignore(filepath.Join(dir, e.Name())) {
			continue
		}
		s.Push(filepath.Join(dir, e.Name()))
	}
	return nil
}

// calcPackagesStat traverses given directory -- excluding the ones for
// which ignore is true -- and adds a given directory as a testing
// package's package stats iff it contains at least one *_test.go file
// which contains at least one test function.  The ModTime-property is
// the modification time of the most recently modified package.
func calcPackagesStat(
	moduleDir, dir string, ignore func(string) bool,
) *packagesStat {

	stk := dirStack{dir}
	pp := packagesStat{}

	for len(stk) > 0 {
		d := stk.Pop()
		if err := stk.PushDir(d, ignore); err != nil {
			continue
		}
		tp, ok := newTestingPackageStat(d)
		if !ok {
			continue
		}
		tp.rel = strings.TrimLeft(
			strings.TrimPrefix(tp.abs, moduleDir), string(os.PathSeparator))
		pp.pp = append(pp.pp, tp)
		if tp.ModTime.After(pp.ModTime) {
			pp.ModTime = tp.ModTime
		}
	}

	return &pp
}

// newTestingPackageStat evaluates if given directory has at least one
// go test file with at least one test in which case a new *pkgStat
// instance and true is returned; otherwise nil and false.  NOTE the
// testing package's modification time is determined as the most recent
// modification time of all *.go source files in this package.  I.e. if
// a package contains none go source files their update have no effect
// on the testing package's modification time.
func newTestingPackageStat(dir string) (*pkgStat, bool) {

	stt, err := os.Stat(dir)
	if err != nil || !stt.IsDir() {
		return nil, false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, false
	}

	stat, testing := &pkgStat{abs: dir}, false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		s, err := os.Stat(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if stat.ModTime.Before(s.ModTime()) {
			stat.ModTime = s.ModTime()
		}
		if testing {
			continue
		}
		if !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		if !isTesting(filepath.Join(dir, e.Name())) {
			continue
		}
		testing = true
	}
	if !testing {
		return nil, false
	}

	stat.loadTestFiles() // TODO: handle error
	return stat, true
}

// isTesting returns true if given file contains at least one function
// declaration (without receiver) whose name is prefixed with "Test";
// otherwise false.
func isTesting(file string) bool {
	fl, err := parser.ParseFile(token.NewFileSet(), file, nil, 0)
	if err != nil {
		return false
	}
	found := false

	ast.Inspect(fl, func(n ast.Node) bool {
		if _, ok := n.(*ast.File); ok {
			return true
		}
		fDcl, ok := n.(*ast.FuncDecl)
		if !ok || found || fDcl.Recv != nil {
			return false
		}
		if !strings.HasPrefix(fDcl.Name.Name, "Test") {
			return false
		}
		found = true
		return false
	})

	return found
}
