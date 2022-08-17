// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// PackagesDiff reports the differences of a module's packages at two
// different points in time.  PackagesDiff is immutable hence we can
// report an instance by sending a pointer over an according channel.
type PackagesDiff struct {
	last, current *testingPackages
}

// For returns all testing packages which were updated since the last
// reported diff.
func (d *PackagesDiff) For(cb func(*TestingPackage) (stop bool)) {
	for _, tp := range d.current.pp {
		if !d.last.isUpdatedBy(tp) {
			continue
		}
		if cb(tp) {
			return
		}
	}
	if d.last == d.current {
		panic("last became current")
	}
}

func (d *PackagesDiff) ForDel(cb func(*TestingPackage) (stop bool)) {
	if d.last == nil || len(d.last.pp) == 0 {
		return
	}
	for _, tp := range d.last.pp {
		if d.current.has(tp) {
			continue
		}
		if cb(tp) {
			return
		}
	}
}

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
	return false
}

type testingPackages struct {
	ModTime time.Time
	pp      []*TestingPackage
}

func (pp *testingPackages) diff(other *testingPackages) *PackagesDiff {
	d := &PackagesDiff{last: other, current: pp}
	if !d.hasDelta() {
		return nil
	}
	return d
}

// isUpdatedBy returns true iff given testing package is not included in
// receiving testing packages or if its modification time is after the
// modification time of corresponding testing package of receiving
// testing packages.
func (pp *testingPackages) isUpdatedBy(tp *TestingPackage) bool {
	if pp == nil {
		return true
	}
	for _, _tp := range pp.pp {
		if _tp.Rel() != tp.Rel() {
			continue
		}
		if !tp.ModTime.After(_tp.ModTime) {
			return false
		}
	}
	return true
}

// has returns true iff receiving testing packages have a package with
// the same relative name as given testing package.
func (pp *testingPackages) has(tp *TestingPackage) bool {
	for _, _tp := range pp.pp {
		if _tp.Rel() != tp.Rel() {
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

// packagesSnapshot traverses to given directory -- excluding the ones
// for which ignore is true -- and adds a given directory as testing
// package iff it contains at least one *_test.go file which contains at
// least one test function.  The MostResentChange-property is the most
// recently updated testing package's update time.
func packagesSnapshot(
	dir string, ignore func(string) bool,
) *testingPackages {

	stk := dirStack{}
	if err := stk.PushDir(dir, ignore); err != nil || len(stk) == 0 {
		return nil
	}

	pp := testingPackages{}

	for len(stk) > 0 {
		d := stk.Pop()
		if err := stk.PushDir(d, ignore); err != nil {
			continue
		}
		tp, ok := newTestingPackage(d)
		if !ok {
			continue
		}
		tp.rel = strings.TrimLeft(
			tp.abs[len(dir):], string(os.PathSeparator))
		pp.pp = append(pp.pp, tp)
		if tp.ModTime.After(pp.ModTime) {
			pp.ModTime = tp.ModTime
		}
	}

	return &pp
}

// newTestingPackage evaluates if given directory has at least one go
// test file with at least one test in which case a new TestingPackage
// instance and true is returned; otherwise nil and false.  NOTE the
// testing package's modification time is determined as the most recent
// modification time of all *.go source files in this package.  I.e. if
// a package contains none go source files their update have no
// influence on the testing packages modification time.
func newTestingPackage(dir string) (*TestingPackage, bool) {

	stt, err := os.Stat(dir)
	if err != nil || !stt.IsDir() {
		return nil, false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, false
	}

	tp, testing := &TestingPackage{abs: dir}, false
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		s, err := os.Stat(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		if tp.ModTime.Before(s.ModTime()) {
			tp.ModTime = s.ModTime()
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

	return tp, true
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
