// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

func reportMixedTests(p *pkg) []interface{} {
	panic("not implemented")
}

// goReport reports tests of a package only containing go tests.
type mxReport struct {
	report
	foldInfo map[uint]suiteInfo
	typ      reportType
}

func (r *mxReport) Type() reportType { return r.typ }

func (r *mxReport) setType(t reportType) { r.typ = t }

func (r *mxReport) Folded() reporter {
	return r
}
