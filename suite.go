// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"go/parser"
	"go/token"
	"reflect"
	"runtime"
	"testing"
)

// Suite implements the private methods of the SuiteEmbedder interface.
// I.e. if you want to run the tests of your own test-suite using
// *gounit.Run* you must embed this type, e.g.:
//
// 		type MySuite struct { gounit.Suite }
//      func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t) }
type Suite struct {
	t     *testing.T
	file  string
	self  interface{}
	value reflect.Value
	rtype reflect.Type
}

func (s *Suite) init(self interface{}, t *testing.T) *Suite {
	s.self, s.t = self, t
	s.value = reflect.ValueOf(self)
	s.rtype = reflect.TypeOf(self)
	_, file, _, ok := runtime.Caller(2)
	if !ok {
		panic("can't determine test-suites file")
	}
	s.file = file
	return s
}

// File returns the file of Run's caller which is typically the
// file-name of the embedding suite.  Add your own File-method to your
// suite to provide a different file.
func (s *Suite) File() string { return s.file }

var special = "Logger"

// run executes all public methods of embedding test-suite which have
// exactly two arguments.
func (s *Suite) run(t *testing.T, indices func(string, string) int) {
	subTestFactory := newSubTestFactory(s)
	for i := 0; i < s.rtype.NumMethod(); i++ {
		method := s.rtype.Method(i)
		if method.Type.NumIn() != 2 {
			continue
		}
		t.Run(method.Name, subTestFactory(
			method,
			indices(s.rtype.Elem().Name(), method.Name),
		))
	}
}

func newIndices(fileName string) func(string, string) int {
	return func(suite, test string) int {
		return indexer.get(fileName, suite, test)
	}
}

// ensureIndexing figures the suite's test-file which is Run and makes
// sure *indexer* has indexed all suite-methods of this file in order of
// their appearance.
func ensureIndexing(suite SuiteEmbedder) (indices func(string, string) int) {
	fSet := token.NewFileSet()
	astFl, err := parser.ParseFile(fSet, suite.File(), nil, 0)
	if err != nil {
		panic(err)
	}
	indexer.ensureIndexingOf(astFl, suite.File())
	return newIndices(suite.File())
}

// SuiteEmbedder is automatically implemented by embedding a
// Suite-instance.  I.e.:
// 		type MySuite struct{ gounit.Suite }
// implements the SuiteEmbedder-interface's private methods.
type SuiteEmbedder interface {
	init(interface{}, *testing.T) *Suite
	run(*testing.T, func(string, string) int)
	File() string
}

// Run sets up embedded Suite-instance and runs all methods of given
// test-suite embedder which are public and have exactly two arguments.
// NOTE the reflection of suite-embedder methods could be more specific,
// i.e. the second argument must be of type *gounit.T*.  To keep
// generated overhead at a minimum all methods with exactly two
// arguments are considered tests.
func Run(suite SuiteEmbedder, t *testing.T) {
	s := suite.init(suite, t)
	indices := ensureIndexing(suite)
	s.run(t, indices)
}

// SuiteLogging implementation of a suite-embedder overwrites provided
// logging mechanism of gounit.T-instances passed to suite-tests with
// provided function of the Logger-method. E.g.:
//
// 		type MySuite {
//	 		gounit.Suite
//			Logs string
// 		}
//
//		func (s *MySuite) Logger() func(...interface{}) {
//			return func(args ...interface{}) {
//				s.Logs += fmt.Sprint(args...)
// 			}
// 		}
//
// 		func (s *MySuite) A_test(t *gounit.T) {
// 			t.Log("A_test has run")
// 		}
//
//		func TestMySuite(t *testing.T) {
//			testSuite := &MySuite{}
//			gounit.Run(testSuite, t)
//			t.Log(testSuite.Logs) // prints "A_test has run" if verbose
// 		}
type SuiteLogging interface {
	Logger() func(args ...interface{})
}

// newSubTestFactory returns for given suite a sub-test-factory, i.e. a
// function wrapping test-methods into function that can be passed to
// the Run-method of a *testing.T*-instance.
func newSubTestFactory(
	suite *Suite,
) func(reflect.Method, int) func(*testing.T) {
	suiteLogging, hasLogger := suite.self.(SuiteLogging)
	return func(test reflect.Method, idx int) func(*testing.T) {
		return func(t *testing.T) {
			suiteT := &T{Idx: idx, t: t, logger: t.Log}
			if hasLogger {
				suiteT.logger = suiteLogging.Logger()
			}
			suiteTVl := reflect.ValueOf(suiteT)
			test.Func.Call([]reflect.Value{suite.value, suiteTVl})
		}
	}
}
