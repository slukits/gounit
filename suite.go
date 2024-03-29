// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"reflect"
	"runtime/debug"
	"strings"
	"testing"
)

// Suite implements the private methods of the SuiteEmbedder interface.
// I.e. if you want to run the tests of your own test-suite using
// [Run] you must embed this type, e.g.:
//
//	type MySuite struct { gounit.Suite }
//
//	// optional Init-method
//	func (s *MySuite) Init(t *gounit.S) { //... }
//
//	// optional SetUp-method
//	func (s *MySuite) SetUp(t *gounit.T) { //... }
//
//	// optional TearDown-method
//	func (s *MySuite) TearDown(t *gounit.T) { //... }
//
//	// ... the suite-tests as public methods of *MySuite ...
//	func (s *MySuite) MyTest(t *gounit.T) { //... }
//
//	// optional Finalize-method
//	func (s *MySuite) Finalize(t *gounit.S) { //... }
//
//	func TestMySuite(t *testing.T) { gounit.Run(&MySuite{}, t) }
type Suite struct {
	t               *testing.T
	self            interface{}
	value           reflect.Value
	rType           reflect.Type
	setUp, tearDown *reflect.Method
}

// newFinalizer returns a function which may be used to register at
// t.Cleanup which calls suite's (given) Finalize-method with provided
// values.
func newFinalizer(
	method *reflect.Method, suite, gounitF reflect.Value,
) func() {
	return func() {
		method.Func.Call([]reflect.Value{suite, gounitF})
	}
}

// exec executes a found Init-method in a Suite.
func (s *Suite) exec(init *reflect.Method, t *testing.T) {
	suiteLogging, hasLogger := s.self.(SuiteLogger)
	suiteCanceler, hasCanceler := s.self.(SuiteCanceler)
	suiteI := &S{
		t:        t,
		logger:   t.Log,
		canceler: t.FailNow,
		prefix:   InitPrefix,
	}
	if hasLogger {
		suiteI.logger = suiteLogging.Logger()
	}
	if hasCanceler {
		suiteI.canceler = suiteCanceler.Cancel()
	}
	init.Func.Call([]reflect.Value{
		s.value, reflect.ValueOf(suiteI)})
}

// sWrapper wraps given testing.T-instance in a S-instance for a suites
// finalizer, i.e. its logging prefix is going to be [FinalPrefix].
func (s *Suite) sWrapper(t *testing.T) *S {
	suiteLogging, hasLogger := s.self.(SuiteLogger)
	suiteCanceler, hasCanceler := s.self.(SuiteCanceler)
	suiteT := &S{
		t:        t,
		logger:   t.Log,
		canceler: t.FailNow,
		prefix:   FinalPrefix,
	}
	if hasLogger {
		suiteT.logger = suiteLogging.Logger()
	}
	if hasCanceler {
		suiteT.canceler = suiteCanceler.Cancel()
	}
	return suiteT
}

// init initializes this suite's reused reflection values and handles
// its special methods if any.
func (s *Suite) init(self interface{}, t *testing.T) *Suite {
	s.self, s.t = self, t
	s.value = reflect.ValueOf(self)
	s.rType = reflect.TypeOf(self)
	for i := 0; i < s.rType.NumMethod(); i++ {
		m := s.rType.Method(i)
		switch m.Name {
		case "SetUp":
			s.setUp = &m
		case "TearDown":
			s.tearDown = &m
		case "Init":
			s.exec(&m, t)
		case "Finalize":
			t.Cleanup(newFinalizer(
				&m, s.value, reflect.ValueOf(s.sWrapper(t))))
		}
	}
	return s
}

const special = "SetUpTearDownInitFinalizeDelGet"

// SuiteEmbedder is automatically implemented by embedding a
// Suite-instance.  I.e.:
//
//	type MySuite struct{ gounit.Suite }
//
// implements the SuiteEmbedder-interface's private methods.
type SuiteEmbedder interface {
	init(interface{}, *testing.T) *Suite
}

// Run sets up embedded Suite-instance and runs all methods of given
// test-suite embedder which are public, have exactly one argument and
// are not special:
//   - Init(*[gounit.S]): run before any other method of a suite
//   - SetUp(*[gounit.T]): run before every suite-test
//   - TearDown(*[gounit.T]): run after every suite-test
//   - Finalize(*[gounit.S]): run after any other method of a suite
//   - Get, Set, Del as methods of [gounit.Fixtures] are also
//     considered special for the use case that Fixtures is
//     embedded in a Suite-embedder (i.e. test-suite)
func Run(suite SuiteEmbedder, t *testing.T) {
	s := suite.init(suite, t)
	subTestFactory := newSubTestFactory(s)
	for i := 0; i < s.rType.NumMethod(); i++ {
		method := s.rType.Method(i)
		if method.Type.NumIn() != 2 {
			continue
		}
		if strings.Contains(special, method.Name) {
			continue
		}
		t.Run(method.Name, subTestFactory(method))
	}
}

// SuiteLogger implementation of a suite-embedder replaces the default
// logging mechanism of [gounit.T]-instances E.g.:
//
//	type MySuite {
//	    gounit.Suite
//	    Logs string
//	}
//
//	func (s *MySuite) Logger() func(...interface{}) {
//	    return func(args ...interface{}) {
//	        s.Logs += fmt.Sprint(args...)
//	    }
//	}
//
//	func (s *MySuite) A_test(t *gounit.T) {
//	    t.Log("A_test has run")
//	}
//
//	func TestMySuite(t *testing.T) {
//	    testSuite := &MySuite{}
//	    gounit.Run(testSuite, t)
//	    t.Log(testSuite.Logs) // prints "A_test has run" if verbose
//	}
type SuiteLogger interface {
	Logger() func(args ...interface{})
}

// SuiteErrorer overwrites default test-error handling which defaults to
// a testing.T.Error-call of a wrapped testing.T-instance.  I.e. calling
// on a [gounit.T] instance t methods like [T.Error] or [T.Errorf] end
// up in an Error-call of the testing.T-instance which is wrapped by t.
// If a suite implements the SuiteErrorer-interface provided function is
// called in case of an test-error.
type SuiteErrorer interface {
	Error() func(...interface{})
}

// SuiteCanceler overwrites default test-cancellation handling which
// defaults to a testing.T.FailNow-call of a wrapped
// testing.T-instance.  I.e. calling on a [gounit.T] instance t
// methods like [T.Fatal], [T.Fatalf], [T.FailNow], [T.FatalIfNot], or
// [T.FatalOn] end up in an FailNow-call of the testing.T-instance which
// is wrapped by t.  If a suite implements the SuiteCanceler-interface
// provided function is called in case of an test-cancellation.
type SuiteCanceler interface {
	Cancel() func()
}

// newSubTestFactory returns for given suite a sub-test-factory, i.e. a
// function wrapping test-methods into function that can be passed to
// the Run-method of a *testing.T*-instance.
func newSubTestFactory(
	suite *Suite,
) func(reflect.Method) func(*testing.T) {
	suiteLogging, hasLogger := suite.self.(SuiteLogger)
	suiteErrorer, hasErrorer := suite.self.(SuiteErrorer)
	suiteCanceler, hasCanceler := suite.self.(SuiteCanceler)
	var tearDown func(t *T)
	if suite.tearDown != nil {
		tearDown = func(t *T) {
			(*suite.tearDown).Func.Call(
				[]reflect.Value{suite.value, reflect.ValueOf(t)})
		}
	}
	return func(test reflect.Method) func(*testing.T) {
		return func(t *testing.T) {
			suiteT := &T{
				t:        t,
				tearDown: tearDown,
				logger:   t.Log,
				errorer:  t.Error,
				canceler: t.FailNow,
			}
			suiteT.Not = Not{t: suiteT}
			if hasLogger {
				suiteT.logger = suiteLogging.Logger()
			}
			if hasErrorer {
				suiteT.errorer = suiteErrorer.Error()
			}
			if hasCanceler {
				suiteT.canceler = suiteCanceler.Cancel()
			}
			suiteTVl := reflect.ValueOf(suiteT)
			if suite.setUp != nil {
				(*suite.setUp).Func.Call(
					[]reflect.Value{suite.value, suiteTVl})
			}
			func(vl []reflect.Value) {
				defer func() {
					if r := recover(); r != nil {
						t.Helper()
						t.Errorf("panicked:\n%v\n%v", r, string(debug.Stack()))
					}
				}()
				test.Func.Call(vl)
			}([]reflect.Value{suite.value, suiteTVl})
			if tearDown != nil {
				tearDown(suiteT)
			}
		}
	}
}
