// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"go/ast"
	"go/token"
	"path/filepath"
	"regexp"
	"strings"
)

type testAst struct {
	fIdx  int
	fs    *token.FileSet
	af    *ast.File
	guSlc string
}

const gounitPath = `"github.com/slukits/gounit"`

// parseTestNSuites parses given ast file for tests and suites and
// associates them with parsed test file.  The parsed tests and suites
// should be used to retrieve results of a test run and the association
// with test file makes it possible to report suites according to their
// associated file's modification time.
func parseTestNSuites(
	fIdx int, fs *token.FileSet, af *ast.File, guSlc string,
) (tt tests, ss suites) {

	ast.Inspect(af, func(n ast.Node) bool {
		if _, ok := n.(*ast.File); ok {
			return true
		}
		fDcl, ok := n.(*ast.FuncDecl)
		if !ok || fDcl.Recv != nil {
			return false
		}
		name, ok := isTest(fDcl)
		if !ok {
			return false
		}
		suite, ok := isSuiteRunner(fDcl, guSlc)
		if !ok {
			tt.add(fIdx, fs.Position(fDcl.Pos()).String(), name)
			return false
		}
		ss.add(fIdx, fs.Position(fDcl.Pos()).String(), suite, name)
		return false
	})

	return tt, ss
}

func parseSuiteTests(ff []*testAst, ss suites) {
	for _, tf := range ff {
		ast.Inspect(tf.af, func(n ast.Node) bool {
			if _, ok := n.(*ast.File); ok {
				return true
			}
			fDcl, ok := n.(*ast.FuncDecl)
			if !ok || fDcl.Recv == nil {
				return false
			}
			suite, test, ok := isSuiteTest(fDcl, ss)
			if !ok {
				return false
			}
			ss.addTest(suite, &Test{
				fIdx: tf.fIdx,
				name: test,
				pos:  int(fDcl.Pos()),
				abs:  tf.fs.Position(fDcl.Pos()).String(),
			})
			return false
		})
	}
}

// parseGounitSelect figures if there is no selector
//
//	import . "github.com/slukits/gounit"
//
// the default selector
//
//	import "github.com/slukits/gounit"
//
// or some other selector to reference gounit's Suite type
//
//	import gu "github.com/slukits/gounit"
func parseGounitSelector(af *ast.File) string {
	for _, i := range af.Imports {
		if i.Path.Value == gounitPath {
			if i.Name != nil {
				if i.Name.Name != "." {
					return i.Name.Name
				}
				return ""
			}
			return filepath.Base(strings.Trim(i.Path.Value, `"`))
		}
	}
	return ""
}

func isTest(fd *ast.FuncDecl) (string, bool) {
	if fd.Recv != nil {
		return "", false
	}
	if !strings.HasPrefix(fd.Name.Name, "Test") {
		return "", false
	}
	return fd.Name.Name, true
}

func isSuiteRunner(
	td *ast.FuncDecl, guSlc string,
) (string, bool) {

	isUnselectedRunner := func(ce *ast.CallExpr) bool {
		if ident, ok := ce.Fun.(*ast.Ident); ok {
			if ident.Name == "Run" {
				return true
			}
		}
		return false
	}

	isSelectedRunner := func(ce *ast.CallExpr) bool {
		slcExp, ok := ce.Fun.(*ast.SelectorExpr)
		if !ok {
			return false
		}
		if slcExp.Sel.Name != "Run" {
			return false
		}
		if ident, ok := slcExp.X.(*ast.Ident); ok {
			if ident.Name == guSlc {
				return true
			}
		}
		return false
	}

	getSuite := func(cd *ast.CallExpr) string {
		ident, ok := cd.Args[0].(*ast.Ident)
		if !ok {
			return cd.Args[0].(*ast.UnaryExpr).X.(*ast.CompositeLit).
				Type.(*ast.Ident).Name
		}
		return ident.Name
	}

	runsSuite, suiteRun := false, ""

	setSuiteRunner := func(s string) {
		runsSuite = true
		suiteRun = s
	}

	ast.Inspect(td, func(n ast.Node) bool {
		switch ce := n.(type) {
		case *ast.CallExpr:
			if guSlc == "" {
				if isUnselectedRunner(ce) {
					setSuiteRunner(getSuite(ce))
				}
				return true
			}
			if !isSelectedRunner(ce) {
				return true
			}
			setSuiteRunner(getSuite(ce))
		}
		return true
	})

	return suiteRun, runsSuite
}

var reUpper = regexp.MustCompile(`^[A-Z]`)
var special = map[string]bool{
	"Init":     true,
	"SetUp":    true,
	"TearDown": true,
	"Finalize": true,
}

// isSuiteTest returns a suite's name, the suite-test's name and true in
// case given function declaration represents a suite-test; zero-strings
// and false otherwise.
func isSuiteTest(fd *ast.FuncDecl, ss suites) (
	string, string, bool) {

	// method with one argument next to its receiver
	if fd.Recv == nil || len(fd.Type.Params.List) != 1 {
		return "", "", false
	}
	// which is neither special nor private
	if special[fd.Name.Name] || !reUpper.MatchString(fd.Name.Name) {
		return "", "", false
	}
	for _, field := range fd.Recv.List {
		name, ok := isIdent(field.Type)
		if !ok {
			continue
		}
		if !ss.has(name) {
			continue
		}
		return name, fd.Name.Name, true
	}
	return "", "", false
}

// isIdent helps investigating if a function's receiver field type
// refers to a known test-suite by returning given field-type's
// identifier-name if their is any.
func isIdent(fldType ast.Expr) (string, bool) {
	if ident, ok := fldType.(*ast.Ident); ok {
		return ident.Name, true
	}

	starExpr, ok := fldType.(*ast.StarExpr)
	if !ok {
		return "", false
	}
	ident, ok := starExpr.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	return ident.Name, true
}
