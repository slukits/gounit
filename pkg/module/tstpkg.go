// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package module

import (
	"context"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// A TestingPackage provides information on a module's package's tests
// and test suites.  As well as the feature to execute and report on a
// package's tests.
type TestingPackage struct {
	ModTime time.Time
	abs, id string
	timeout time.Duration
	parsed  bool
	files   []*testFile
	tests   tests
	suites  suites
}

// Name returns the testing package's name.
func (tp TestingPackage) Name() string { return filepath.Base(tp.abs) }

// Abs returns the absolute path *to* the testing package, i.e. Abs
// doesn't include the packages name.
func (tp TestingPackage) Abs() string { return filepath.Dir(tp.abs) }

// Rel returns the module relative path *to* the testing package, i.e. Rel
// doesn't include the packages name.
func (tp TestingPackage) Rel() string { return filepath.Dir(tp.id) }

// ID returns the module-relative package path including the package's
// name.  Hence ID() is a module-global unique identifier of given
// package.
func (tp TestingPackage) ID() string { return tp.id }

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

const StdErr = "shell exit error: "

// Run executes go test for the testing package an returns its result.
// Returned error if any is the error of command execution, i.e. a
// timeout.  While Result.Err reflects errors from the error console.
// Note right before the Run executes its command the testing package's
// test files are parsed to increase the likeliness that the parsed
// result matches the ran tests.  I.e. time is saved if tests and test
// suites are accessed after the call of Run.  The output of the go
// testing tool is sadly not enough to report tests in the order they
// were written if tests run concurrently.  Hence to achieve the goal
// that the test reporting outlines the documentation and thought
// process of the production code, i.e. tests are reported in the order
// they were written, it is necessary to parse the test files separately
// and then match the findings to the result of the test run.
func (tp *TestingPackage) Run() (*Results, error) {
	ctx, cancel := context.WithTimeout(
		context.Background(), tp.timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "test", "-json")
	cmd.Dir = tp.abs
	start := time.Now()
	stdout, err := cmd.CombinedOutput()
	if err != nil {
		if _, ok := err.(*exec.ExitError); !ok {
			return nil, fmt.Errorf("%s: %s", cmd.String(), string(stdout))
		}
	}
	duration := time.Since(start)
	rr, jsonErr := unmarshal(stdout)
	if jsonErr != nil {
		if err != nil {
			return &Results{duration: time.Since(start),
				err: fmt.Sprintf("%s%v: %s",
					StdErr, err, string(stdout))}, nil
		}
		return &Results{duration: time.Since(start),
			err: fmt.Sprintf("json-unmarshal stdout: %v", err)}, nil
	}
	return &Results{rr: rr, duration: duration}, nil
}

type testAst struct {
	fs    *token.FileSet
	af    *ast.File
	name  string
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

	ff := []*testAst{}
	tt, ss := tests{}, suites{}
	for _, e := range ee {
		if e.IsDir() || !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		fs := token.NewFileSet()
		flName := filepath.Join(tp.abs, e.Name())
		af, err := parser.ParseFile(fs,
			flName, nil, 0)
		if err != nil {
			continue
		}
		guSlc := parseGounitSelector(af)
		ff = append(ff, &testAst{
			fs: fs, af: af, name: flName, guSlc: guSlc})
		_tt, _ss := parseTestNSuites(fs, af, guSlc)
		tt, ss = append(tt, _tt...), append(ss, _ss...)
	}
	parseSuiteTests(ff, ss)
	tp.tests = tt
	tp.suites = ss
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

// A Test provides information about a go test, i.e. Test*-function.
type Test struct {
	file string
	name string
	pos  int
	abs  string
}

// Name returns a tests name.
func (t *Test) Name() string { return t.name }

// Pos returns a tests absolute filename with line and column number.
func (t *Test) Pos() string { return t.abs }

type tests []*Test

func (tt *tests) add(pos, name string) {
	*tt = append(*tt, &Test{abs: pos, name: name})
}

type TestSuite struct {
	Test
	runner string
	tests  []*Test
}

// Runner returns the Test*-function's name which is executing given
// test suite.
func (s *TestSuite) Runner() string { return s.runner }

// ForTest provides given test suite's tests.
func (s *TestSuite) ForTest(cb func(*Test)) {
	for _, t := range s.tests {
		cb(t)
	}
}

type suites []*TestSuite

func (ss *suites) add(pos, name, runner string) {
	*ss = append(*ss, &TestSuite{
		Test:   Test{abs: pos, name: name},
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

// A pkgStat is calculated to determine if a package changed in the
// course of time.
type pkgStat struct {
	ModTime  time.Time
	abs, rel string
}

// Name returns a testing package's name.
func (ps pkgStat) Name() string { return filepath.Base(ps.abs) }

// Abs returns the absolute path *to* the stated testing package, i.e.
// Abs doesn't include the packages name.
func (ps pkgStat) Abs() string { return filepath.Dir(ps.abs) }

// ID returns the module-relative path including the package itself.
// I.e. ID() is a module-global unique identifier of a testing package.
func (ps pkgStat) ID() string { return ps.rel }

// loadTestFiles reads the test-files of a testing package to order
// parsed suits by modification time of their associated test files.
func (ps *pkgStat) loadTestFiles() (tt []*testFile, err error) {
	ee, err := os.ReadDir(ps.abs)
	if err != nil {
		return nil, err
	}

	for _, e := range ee {
		if !strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		stt, err := os.Stat(filepath.Join(ps.abs, e.Name()))
		if err != nil {
			return nil, err
		}
		bb, err := os.ReadFile(filepath.Join(ps.abs, e.Name()))
		if err != nil {
			return nil, err
		}
		tt = append(
			tt, &testFile{
				modTime: stt.ModTime(),
				name:    filepath.Join(ps.abs, e.Name()),
				content: bb,
			})
	}
	return tt, nil
}

type testFile struct {
	modTime time.Time
	name    string
	content []byte
}
