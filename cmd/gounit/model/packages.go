// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
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
	pp := []*TestingPackage{}
	for _, ps := range d.current.pp {
		if !d.last.isUpdatedBy(ps) {
			continue
		}
		tt, err := ps.loadTestFiles()
		if err != nil {
			return err
		}
		id := ps.rel
		if id == "" {
			id = filepath.Base(ps.abs)
		}
		pp = append(pp, &TestingPackage{
			ModTime: ps.ModTime,
			id:      id, abs: ps.abs, files: tt, Timeout: d.timeout})
	}
	sort.Slice(pp, func(i, j int) bool {
		return pp[i].ModTime.After(pp[j].ModTime)
	})
	for _, p := range pp {
		if cb(p) {
			return nil
		}
	}
	return nil
}

func (d *PackagesDiff) String() string {
	ss := []string{}
	dd := []*TestingPackage{}
	d.For(func(tp *TestingPackage) (stop bool) {
		dd = append(dd, tp)
		return
	})
	switch len(dd) {
	case 0:
		ss = append(ss, "pkg-diff: updated: none")
	default:
		ss = append(ss, "pkg-diff: updated:")
		for _, d := range dd {
			ss = append(ss, fmt.Sprintf("  %s", d.ID()))
		}
	}
	dd = []*TestingPackage{}
	d.ForDel(func(tp *TestingPackage) (stop bool) {
		dd = append(dd, tp)
		return
	})
	switch len(dd) {
	case 0:
		ss = append(ss, "pkg-diff: deleted: none")
	default:
		ss = append(ss, "pkg-diff: deleted:")
		for _, d := range dd {
			ss = append(ss, fmt.Sprintf("  %s", d.ID()))
		}
	}
	return strings.Join(ss, "\n")
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
		id := ps.rel
		if id == "" {
			id = filepath.Base(ps.abs)
		}
		tp := &TestingPackage{
			id: id, abs: ps.abs, parsed: true, Timeout: 0}
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
