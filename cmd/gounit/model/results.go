// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/slukits/gounit"
)

// Results reports the results for each go Test* function of a testing
// package's test run.  Results of sub tests are reported by their
// parent test.  Results may be queried leveraging a testing package's
// parsed tests, test suites and suite tests. E.g. let pkg be a
// TestingPackage instance reported to a module watcher.
//
//	rr, err := pkg.Run()
//	panic(err) // before executed "go test" command finished
//	if rr.HasErr() { // from stderr after command execution finished
//	        panic(rr.Err())
//	}
//	pkg.ForTest(func(t *module.Test) {
//	    fmt.Printf("%s passed: %v\n", t.Name(), rr.OfTest(t).Passed)
//	})
//	pkg.ForSuite(func(ts *module.TestSuite) {
//	    sr := rr.OfSuite(ts)
//	    fmt.Printf("suite %s passed: %v", ts.Name(), rr.OfSuite(ts).Passed)
//	    ts.ForTest(func(t *module.Test) {
//	        fmt.Printf("\t%s passed: %v\n", t.Name(), sr.Of(t).Passed)
//	    })
//	})
type Results struct {

	// rr holds the results of a testing package's test run
	rr results

	// Duration of a test run.
	Duration time.Duration

	// err from the error console
	err string
}

func newResult(stdout []byte) {}

// Err reports a shell exit error of a tests run.
func (r *Results) Err() string { return r.err }

// HasErr returns true if a tests run resulted in a shell exit error.
func (r *Results) HasErr() bool { return r.err != "" }

// OfTest returns the test result of given Test instance representing a
// go Test* function (which is not running a test-suite).
func (r *Results) OfTest(t *Test) *TestResult { return r.rr[t.Name()] }

// OfSuite returns the test result of given test suite and its suite
// tests.
func (r *Results) OfSuite(ts *TestSuite) *TestResult {
	return r.rr[ts.Runner()]
}

func (r *Results) Passed() bool {
	for _, _r := range r.rr {
		if _r.Passed {
			continue
		}
		return false
	}
	return true
}

// Len reports the number of tests, i.e. the number of go Test* tests
// plus the suite runners.  Results has no option to distinguish suite
// "runners" from "normal" go Test* tests.  For this the parsed suite
// information of a testing package needs to be leveraged.
func (r *Results) Len() int { return len(r.rr) }

// Result instance is embedded in a TestResult or SubResult and
// expresses their commonalities.  There are two result types needed
// because a TestResult may represent a test suite which in turn may
// report test logs of the suites Init- or Finalize-method.  While
// SubResult instances can't have this.
type Result struct {
	Passed  bool
	Skipped bool
	Panics  bool
	inRace  bool
	Output  []string
	Start   time.Time
	End     time.Time
	Name    string
	subs    subResults
}

func (r *Result) panicErr() string {
	return strings.Join(append([]string{r.Name}, r.Output...), "\n")
}

// Len is the number of executed test comprising given test result.
// I.e. it is either 1 given result has no sub test results or the
// number of executed sub tests.  I.e. tests having sub tests are not
// counted.
func (r *Result) Len() int {
	if len(r.subs) == 0 {
		return 1
	}
	n := 0
	for _, s := range r.subs {
		n += s.Len()
	}
	return n
}

// HasSubs allows to discriminate go-tests with one sub-test from a
// single go-test.
func (r *Result) HasSubs() bool { return len(r.subs) > 0 }

func (r *Result) String() string {
	name := r.Name
	if strings.Contains(r.Name, "_") {
		name = strings.ReplaceAll(r.Name, "_", " ")
		for i, c := range name {
			name = string(unicode.ToLower(c)) + name[i+1:]
			break
		}
	}
	return apostrophe(camelCaseToHuman(name))
}

// LenFailed returns the number of failed tests which is only
// interesting in case of sub results otherwise a Result's Passed
// property could be consulted.
func (r *Result) LenFailed() int {
	if len(r.subs) == 0 {
		if r.Passed {
			return 0
		}
		return 1
	}
	n := 0
	for _, s := range r.subs {
		n += s.LenFailed()
	}
	return n
}

// For calls back for each sub test result of a test result.  I.e. in
// case of a suite runner for each suite test.  Since it never
// occurred to me to nest tests deeper than that the support for this
// use case is rather rudimentary see [result.Descend].
func (r *Result) For(cb func(*SubResult)) {
	for _, s := range r.subs {
		cb(s)
	}
}

// For calls back for each sub test result of a test result.  I.e. in
// case of a suite runner for each suite test.
func (r *Result) ForOrdered(cb func(*SubResult)) {
	sort.Slice(r.subs, func(i, j int) bool {
		return r.subs[i].Name < r.subs[j].Name
	})
	for _, s := range r.subs {
		cb(s)
	}
}

// Sub returns the result of the sub test with given name.
func (r *Result) OfTest(t *Test) *SubResult {
	for _, sr := range r.subs {
		if sr.Name != t.Name() {
			continue
		}
		return sr
	}
	return nil
}

// Descend provides a depth first traversing of a sub test result having
// itself sub test results and so on.
func (r *Result) Descend(sr *SubResult, cb func(parent, sr *SubResult)) {
	sr.For(func(_sr *SubResult) {
		cb(sr, _sr)
		_sr.Descend(_sr, cb)
	})
}

// TestResult indicates if a test has passed and what output it has
// generated.
type TestResult struct {
	*Result

	// InitOut reports the output of a test suites Init-method.
	InitOut []string

	// FinalizeOut reports the output of a test suites Finalize-method.
	FinalizeOut []string
}

type subResults []*SubResult

func (sr *subResults) get(test string) *SubResult {
	for _, sr := range *sr {
		if sr.Name != test {
			continue
		}
		return sr
	}
	return nil
}

func (sr *subResults) add(test string) *SubResult {
	_sr := &SubResult{Result: &Result{Name: test}}
	*sr = append(*sr, _sr)
	return _sr
}

// A SubResult of a run sub test is reported by a Result instance r:
//
//	r.For(func(sr *SubResult) {
//	    // do some thing with sub test result
//	})
type SubResult struct {
	*Result
}

const (
	acRun    = "run"    // the test has started running
	acPause  = "pause"  // the test has been paused
	acCont   = "cont"   // the test has continued running
	acPass   = "pass"   // the test passed
	acBench  = "bench"  // benchmark printed log output but did not fail
	acFail   = "fail"   // the test or benchmark failed
	acOutput = "output" // the test printed output
	acSkip   = "skip"   // test was skipped or package contained no tests
)

type event struct {
	Time    time.Time // encodes as an RFC3339-format string
	Action  string
	Package string
	Test    string
	Elapsed float64 // seconds
	Output  string
}

// jsonProperties must be all present in a provided stdout in order to
// unmarshal to Events.
var jsonProperties = [][]byte{
	[]byte("Time"), []byte("Action"), []byte("Package"), []byte("Test"),
	[]byte("Output"), []byte("Elapsed")}

func unmarshal(stdout []byte) (results, error) {
	for _, p := range jsonProperties {
		if !bytes.Contains(stdout, p) {
			return nil, fmt.Errorf("unmarshal test-run: "+
				"stdout not parsable:\n%s", string(stdout))
		}
	}
	rr := results{}
	for _, raw := range bytes.Split(bytes.TrimSpace(stdout), []byte("\n")) {
		event := &event{}
		if err := json.Unmarshal(raw, event); err != nil {
			return nil, err
		}
		rr.addEvent(event)
	}
	return rr.passSubSubs(), nil
}

var (
	reSkip = regexp.MustCompile(`^\s*(===|---)`)
)

type results map[string]*TestResult

func (r results) hasPanic() (string, bool) {
	for _, t := range r {
		if t.Panics {
			return t.panicErr(), true
		}
		if !t.HasSubs() {
			continue
		}
		err := ""
		t.For(func(sr *SubResult) {
			if err != "" || (!sr.Panics && !sr.HasSubs()) {
				return
			}
			if sr.Panics {
				err = sr.panicErr()
				return
			}
			sr.For(func(sr *SubResult) {
				if err != "" || !sr.Panics {
					return
				}
				err = sr.panicErr()
			})
		})
		if err == "" {
			continue
		}
		return err, true
	}
	return "", false
}

// passSubSubs it seems that go test passing sub-tests having sub-tests
// them self is not reporting as passing; hence we make them pass if all
// their sub-tests pass.
func (r results) passSubSubs() results {
	for _, t := range r {
		if !t.HasSubs() {
			continue
		}
		t.For(func(sr *SubResult) {
			if !sr.HasSubs() {
				return
			}
			sr.Passed = true
			sr.For(func(s *SubResult) {
				if sr.Passed && s.Passed {
					return
				}
				sr.Passed = false
			})
		})
	}
	return r
}

func (r *results) addEvent(e *event) {
	if e.Test == "" {
		return
	}
	rslt := r.get(e.Test)
	switch e.Action {
	case acRun:
		if rslt.Start.IsZero() || e.Time.Before(rslt.Start) {
			rslt.Start = e.Time
		}
	case acPass:
		rslt.Passed = true
		rslt.End = e.Time
	case acFail:
		rslt.End = e.Time
	case acSkip:
		rslt.Passed = true
		rslt.Skipped = true
	case acOutput:
		if reSkip.MatchString(e.Output) {
			if rslt.inRace {
				rslt.inRace = false
			}
			break
		}
		if (strings.HasPrefix(e.Output, "panic:") ||
			strings.HasPrefix(e.Output, "\tpanic:")) && !rslt.Panics {

			rslt.Panics = true
		}
		if strings.Contains(e.Output, gounit.InitPrefix) {
			tr, ok := (*r)[e.Test]
			if !ok {
				break
			}
			out := strings.Replace(e.Output, gounit.InitPrefix, "", 1)
			for _, s := range strings.Split(out, "\n") {
				if s == "" {
					continue
				}
				tr.InitOut = append(tr.InitOut, strings.TrimSpace(s))
			}
			break
		}
		if strings.Contains(e.Output, gounit.FinalPrefix) {
			tr, ok := (*r)[e.Test]
			if !ok {
				break
			}
			out := strings.Replace(e.Output, gounit.FinalPrefix, "", 1)
			for _, s := range strings.Split(out, "\n") {
				if s == "" {
					continue
				}
				tr.FinalizeOut = append(tr.FinalizeOut, strings.TrimSpace(s))
			}
			break
		}
		for _, s := range strings.Split(e.Output, "\n") {
			if s == "" {
				continue
			}
			rslt.Output = append(rslt.Output, strings.TrimSpace(s))
		}
		if strings.Contains(e.Output, "WARNING: DATA RACE") {
			rslt.inRace = true
			rslt.Passed = false
		}
	}
}

func (r *results) get(testName string) *Result {
	path := strings.SplitN(testName, "/", 3)
	root, ok := (*r)[path[0]]
	if !ok {
		root = &TestResult{Result: &Result{Name: path[0]}}
		(*r)[path[0]] = root
	}
	if len(path) == 1 {
		return root.Result
	}
	rslt := root.subs.get(path[1])
	if rslt == nil {
		rslt = root.subs.add(path[1])
	}
	if len(path) == 2 {
		return rslt.Result
	}
	final := rslt.subs.get(path[2])
	if final == nil {
		return rslt.subs.add(path[2]).Result
	}
	return final.Result
}
