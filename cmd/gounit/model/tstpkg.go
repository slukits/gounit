// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/slukits/ints"
	"golang.org/x/exp/slices"
)

// A TestingPackage provides information on a module's package's tests
// and test suites.  As well as the feature to execute and report on a
// package's tests.
type TestingPackage struct {
	ModTime  time.Time
	abs, id  string
	Timeout  time.Duration
	parsed   bool
	parseErr error
	files    []*testFile
	tests    tests
	suites   suites

	// srcStats caches once calculated source stats since a testing
	// package is re-reported in case of an update.
	srcStats *SrcStats
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

// LenTests returns the number of go tests of a testing package.
func (tp *TestingPackage) LenTests() int {
	if err := tp.ensureParsing(); err != nil {
		return 0
	}
	return len(tp.tests)
}

// HasSrcStats returns true if given testing package has its source
// stats calculation stored.
func (tp *TestingPackage) HasSrcStats() bool {
	return tp.srcStats != nil
}

// ResetSrcStats resets the source stats, i.e. HasSrcStats will return
// false after a call of ResetSrcStats until SrcStats is requested
// again.
func (tp *TestingPackage) ResetSrcStats() {
	tp.srcStats = nil
}

// SrcStats provides statistics about code files/lines and documentation
// of a testing package.
func (tp *TestingPackage) SrcStats() *SrcStats {
	if tp.srcStats == nil {
		tp.srcStats = newSrcStats(tp)
	}
	return tp.srcStats
}

// ForTest provides given testing package's tests.  ForTest fails in
// case of an parse error.
func (tp *TestingPackage) ForTest(cb func(*Test)) error {
	if err := tp.ensureParsing(); err != nil {
		return err
	}
	for _, t := range tp.tests {
		cb(t)
	}
	return nil
}

// LenSuites returns the number of suites of a testing package.
func (tp *TestingPackage) LenSuites() int {
	if err := tp.ensureParsing(); err != nil {
		return 0
	}
	return len(tp.suites)
}

// ForSuite provides given testing package's suites.  ForSuite fails in
// case of an parse error.  Note the last suite is of the package's most
// recently modified test file.
func (tp *TestingPackage) ForSuite(cb func(*TestSuite)) error {
	if err := tp.ensureParsing(); err != nil {
		return err
	}
	for _, s := range tp.suites {
		cb(s)
	}
	return nil
}

// TrimTo removes all parsed tests and suites which are not found in
// given results.  (This may happen if tests are excluded due to build
// tags)
func (tp *TestingPackage) TrimTo(rr *Results) {
	if err := tp.ensureParsing(); err != nil {
		return
	}
	var delTT, delSS []int
	for idx, t := range tp.tests {
		if r := rr.OfTest(t); r != nil {
			continue
		}
		delTT = append(delTT, idx)
	}
	for i, idx := range delTT {
		tp.tests = slices.Delete(tp.tests, idx-i, (idx-i)+1)
	}
	for idx, s := range tp.suites {
		if r := rr.OfSuite(s); r != nil {
			continue
		}
		delSS = append(delSS, idx)
	}
	for i, idx := range delSS {
		tp.suites = slices.Delete(tp.suites, idx-i, (idx-i)+1)
	}
}

// ForSortedSuite calls back for each suite of given package whereas the
// suites are ordered by name instead of the modification date of the
// test file they belong to.
func (tp *TestingPackage) ForSortedSuite(cb func(*TestSuite)) error {
	if err := tp.ensureParsing(); err != nil {
		return err
	}
	if len(tp.suites) == 0 {
		return nil
	}
	sorted := append(suites{}, tp.suites...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].name < sorted[j].name
	})
	for _, s := range sorted {
		cb(s)
	}
	return nil
}

// LastSuite returns the last parsed suite of the most recently modified
// test file in given testing package.
func (tp *TestingPackage) LastSuite() *TestSuite {
	if err := tp.ensureParsing(); err != nil {
		return nil
	}
	if len(tp.suites) == 0 {
		return nil
	}
	return tp.suites[len(tp.suites)-1]
}

// Suite returns the test suite with given name or nil.
func (tp *TestingPackage) Suite(name string) *TestSuite {
	if err := tp.ensureParsing(); err != nil {
		return nil
	}
	if len(tp.suites) == 0 {
		return nil
	}
	for _, s := range tp.suites {
		if s.name != name {
			continue
		}
		return s
	}
	return nil
}

const StdErr = "shell exit error: "

// RunMask controls set flags for a test run.
type RunMask uint8

const (
	// RunVet removes the -vet=off flag from a test run
	RunVet RunMask = 1 << iota
	// RunRace adds the -race flag to a test run
	RunRace
)

// Run executes go test for the testing package and returns its result.
// Returned error if any is the error of command execution, i.e. a
// timeout.  While Result.Err reflects errors from the error console.
// Note the output of the go testing tool is sadly not enough to report
// tests in the order they were written if tests run concurrently.
// Hence to achieve the goal that the test reporting outlines the
// documentation and thought process of the production code, i.e. tests
// are reported in the order they were written, it is necessary to parse
// the test files separately and then match the findings to the result
// of the test run.
func (tp *TestingPackage) Run(rm RunMask) (*Results, error) {
	tp.parsed = false
	ctx, cancel := context.WithTimeout(
		context.Background(), tp.Timeout)
	defer cancel()
	aa := []string{"test", "-json"}
	if rm&RunVet == 0 {
		aa = append(aa, "-vet=off")
	}
	if rm&RunRace != 0 {
		aa = append(aa, "-race")
	}
	aa = append(aa, fmt.Sprintf("-timeout=%s", tp.Timeout))
	cmd := exec.CommandContext(ctx, "go", aa...)
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
			return &Results{Duration: time.Since(start),
				err: fmt.Sprintf("%s%v:\n%s",
					StdErr, err, string(stdout))}, nil
		}
		return &Results{Duration: time.Since(start),
			err: fmt.Sprintf("json-unmarshal stdout: %v", err)}, nil
	}
	if err, ok := rr.hasPanic(); ok {
		return &Results{
			rr: rr, Duration: duration,
			err: err,
		}, nil
	}
	return &Results{rr: rr, Duration: duration}, nil
}

func (tp *TestingPackage) ensureParsing() error {
	if tp.parsed {
		return tp.parseErr
	}
	tp.parsed = true

	ff := []*testAst{}
	tt, ss := tests{}, suites{}
	for idx, tf := range tp.files {
		fs := token.NewFileSet()
		af, err := parser.ParseFile(fs, tf.name, tf.content, 0)
		if err != nil {
			tp.parseErr = err
			return err
		}
		guSlc := parseGounitSelector(af)
		ff = append(ff, &testAst{
			fIdx: idx, fs: fs, af: af, guSlc: guSlc})
		_tt, _ss := parseTestNSuites(idx, fs, af, guSlc)
		tt, ss = append(tt, _tt...), append(ss, _ss...)
	}
	parseSuiteTests(ff, ss)
	ss.sort(tp.files)
	tp.tests = tt
	tp.suites = ss
	return nil
}

// A Test provides information about a go test, i.e. Test*-function.
type Test struct {
	fIdx int
	name string
	pos  int
	abs  string
}

// Name returns a tests name.
func (t *Test) Name() string { return t.name }

var (
	camelRe   = regexp.MustCompile(`\p{Lu}+[0-9.,!\- ]*`)
	endsInNum = regexp.MustCompile(`\p{Lu}+[0-9.,!\- ]+`)
	brokenEnd = regexp.MustCompile(`\p{Lu} \p{Ll}$`)
)

func (t *Test) String() string {
	return HumanReadable(t.name)
}

func HumanReadable(name string) string {
	if strings.Contains(name, "_") {
		name = strings.ReplaceAll(name, "_", " ")
		for i, c := range name {
			name = string(unicode.ToLower(c)) + name[i+1:]
			break
		}
	}
	return apostrophe(camelCaseToHuman(name))
}

func apostrophe(name string) string {
	name = strings.ReplaceAll(name, " s ", "'s ")
	name = strings.ReplaceAll(name, "dont", "don't")
	name = strings.ReplaceAll(name, "doesnt", "doesn't")
	name = strings.ReplaceAll(name, "havent", "haven't")
	name = strings.ReplaceAll(name, "hasnt", "hasn't")
	name = strings.ReplaceAll(name, "isnt", "isn't")
	return name
}

func camelCaseToHuman(str string) string {
	str = strings.TrimSpace(camelRe.ReplaceAllStringFunc(
		str, func(s string) string {
			if len(s) == 1 {
				return " " + strings.ToLower(s)
			}
			if endsInNum.MatchString(s) {
				return " " + s
			}
			return " " + s[:len(s)-1] + " " + strings.ToLower(string(s[len(s)-1]))
		}))
	str = strings.ReplaceAll(str, "  ", " ")
	str = brokenEnd.ReplaceAllStringFunc(str, func(s string) string {
		prefix := rune(0)
		for i, r := range s {
			if i == 0 {
				prefix = r
				continue
			}
			if r == ' ' {
				continue
			}
			return string(prefix) + strings.ToUpper(string(r))
		}
		return s
	})
	return strings.TrimPrefix(str, "test ")
}

// Pos returns a tests absolute filename with line and column number.
func (t *Test) Pos() string { return t.abs }

type tests []*Test

func (tt *tests) add(fIdx int, pos, name string) {
	*tt = append(*tt, &Test{fIdx: fIdx, abs: pos, name: name})
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

func (s *TestSuite) mostRecent(ff []*testFile) (idx int) {

	ii := (&ints.Set{}).Add(s.fIdx)

	for _, t := range s.tests {
		if ii.Has(t.fIdx) {
			continue
		}
		ii.Add(t.fIdx)
	}

	mostRecent := ii.ToSlice()[0]

	if ii.Len() == 1 {
		return mostRecent
	}
	for _, idx := range ii.ToSlice()[1:] {
		if ff[idx].modTime.Before(ff[mostRecent].modTime) {
			continue
		}
		mostRecent = idx
	}
	return mostRecent
}

type suites []*TestSuite

func (ss *suites) add(fIdx int, pos, name, runner string) {
	*ss = append(*ss, &TestSuite{
		Test:   Test{fIdx: fIdx, abs: pos, name: name},
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

func (ss suites) sort(ff []*testFile) {
	sort.Slice(ss, func(i, j int) bool {
		iIdx, jIdx := ss[i].mostRecent(ff), ss[j].mostRecent(ff)
		if iIdx == jIdx {
			return ss[i].pos < ss[j].pos
		}
		less := ff[iIdx].modTime.Before(ff[jIdx].modTime)
		return less
	})
}

type testFile struct {
	modTime time.Time
	name    string
	content []byte
}
