// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

// PackagesDiff reports the differences of a module's packages at two
// different points in time.  PackagesDiff is immutable hence we can
// report an instance by sending a pointer over an according channel.
type PackagesDiff struct{}

type packages map[*TestingPackage]bool

func packagesSnapshot(module string) packages {
	pp := packages{}
	return pp
}

func (pp *packages) diff(other packages) *PackagesDiff {
	return &PackagesDiff{}
}

type TestingPackage struct{}
