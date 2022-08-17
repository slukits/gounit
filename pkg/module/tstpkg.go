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
	"regexp"
	"strings"
	"time"
)

type TestingPackage struct {
	ModTime  time.Time
	abs, rel string
	parsed   bool
	tests    tests
	suites   suites
}

// Name returns the testing package's name.
func (tp TestingPackage) Name() string { return filepath.Base(tp.abs) }

// Abs returns the absolute path to the testing package, i.e. Abs
// doesn't include the packages name.
func (tp TestingPackage) Abs() string { return filepath.Dir(tp.abs) }

// Rel returns the module-relative path including the package itself.
// I.e. Rel() is a module-global unique identifier of given package.
func (tp TestingPackage) Rel() string { return tp.rel }

// ForTest provides given testing package's tests.
func (tp *TestingPackage) ForTest(cb func(*Test)) {
	tp.ensureParsing()
	for _, t := range tp.tests {
		cb(t)
	}
}

// ForSuite provides given testing package's suites.
func (tp *TestingPackage) ForSuite(cb func(*TestSuite)) {
	tp.ensureParsing()
	for _, s := range tp.suites {
		cb(s)
	}
}

type testFile struct {
	fs    *token.FileSet
	af    *ast.File
	guSlc string
}

func (tp *TestingPackage) ensureParsing() {
	if tp.parsed {
		return
	}
	tp.parsed = true
	ee, err := os.ReadDir(tp.abs)
	if err != nil {
		return
	}

	ff := []*testFile{}
	for _, e := range ee {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		fs := token.NewFileSet()
		af, err := parser.ParseFile(fs,
			filepath.Join(tp.abs, e.Name()), nil, 0)
		if err != nil {
			continue
		}
		guSlc := parseGoUnitSelector(af)
		ff = append(ff, &testFile{fs: fs, af: af, guSlc: guSlc})
		tt, ss := parseTestNSuites(fs, af, guSlc)
		tp.tests = append(tp.tests, tt...)
		tp.suites = append(tp.suites, ss...)
	}
	parseSuiteTests(ff, tp.suites)
}

const gounitPath = `"github.com/slukits/gounit"`

func parseTestNSuites(
	fs *token.FileSet, af *ast.File, guSlc string,
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
			tt.add(fs.Position(fDcl.Pos()).String(), name)
			return false
		}
		ss.add(fs.Position(fDcl.Pos()).String(), suite, name)
		return false
	})

	return tt, ss
}

func parseSuiteTests(ff []*testFile, ss suites) {
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
				name: test,
				pos:  tf.fs.Position(fDcl.Pos()).String(),
			})
			return false
		})
	}
}

func parseGoUnitSelector(af *ast.File) string {
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

type Test struct {
	name string
	pos  string
}

func (t *Test) Name() string { return t.name }

func (t *Test) Pos() string { return t.pos }

type tests []*Test

func (tt *tests) add(pos, name string) {
	*tt = append(*tt, &Test{pos: pos, name: name})
}

type TestSuite struct {
	Test
	runner string
	tests  []*Test
}

func (s *TestSuite) Runner() string { return s.runner }

func (s *TestSuite) ForTest(cb func(*Test)) {
	for _, t := range s.tests {
		cb(t)
	}
}

type suites []*TestSuite

func (ss *suites) add(pos, name, runner string) {
	*ss = append(*ss, &TestSuite{
		Test:   Test{pos: pos, name: name},
		runner: runner,
	})
}

func (ss *suites) addTest(suite string, t *Test) {
	for _, s := range *ss {
		if s.name != suite {
			continue
		}
		s.tests = append(s.tests, t)
		return
	}
}

func (ss suites) has(name string) bool {
	for _, s := range ss {
		if s.name != name {
			continue
		}
		return true
	}
	return false
}
