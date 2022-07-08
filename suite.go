// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"reflect"
	"strings"
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
	self  interface{}
	value reflect.Value
	rtype reflect.Type
}

func (s *Suite) init(self interface{}, t *testing.T) *Suite {
	s.self, s.t = self, t
	s.value = reflect.ValueOf(self)
	s.rtype = reflect.TypeOf(self)
	return s
}

var special = "Logger"

// run executes all public methods of embedding test-suite which are not
// special.
func (s *Suite) run(t *testing.T) {
	subTestFactory := newSubTestFactory(s)
	for i := 0; i < s.rtype.NumMethod(); i++ {
		method := s.rtype.Method(i)
		if strings.Contains(special, method.Name) {
			continue
		}
		t.Run(method.Name, subTestFactory(method))
	}
}

// SuiteEmbedder is automatically implemented by embedding a
// Suite-instance.  I.e.:
// 		type MySuite struct{ gounit.Suite }
// implements the SuiteEmbedder-interface's private methods.
type SuiteEmbedder interface {
	init(interface{}, *testing.T) *Suite
	run(*testing.T)
}

// Run runs all methods of given test-suite embedder which are public
// and not special.  Special methods are:
// - Logger: overwriting gounit.T's logging mechanism
func Run(suite SuiteEmbedder, t *testing.T) {
	s := suite.init(suite, t)
	s.run(t)
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
) func(reflect.Method) func(*testing.T) {
	suiteLogging, hasLogger := suite.self.(SuiteLogging)
	return func(test reflect.Method) func(*testing.T) {
		return func(t *testing.T) {
			suiteT := &T{t: t, logger: t.Log}
			if hasLogger {
				suiteT.logger = suiteLogging.Logger()
			}
			suiteTVl := reflect.ValueOf(suiteT)
			test.Func.Call([]reflect.Value{suite.value, suiteTVl})
		}
	}
}
