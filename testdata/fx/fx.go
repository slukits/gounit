// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// Package fx provides gounit test-fixture suites.
//
// Each test-fixture suite embeds the FixtureLog ensuring that all
// loggings during a suite's test runs are appended to the
// *Logs*-property which then can be evaluate after the suite's test
// runs.
package fx

import (
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/slukits/gounit"
)

// A FX instance is meant for fixture suites whose test errors should be
// suppressed and only logged, i.e. may be retrieved by s.Logs where s
// is a suite instance embedding FX.  In this why test-test-suites like
// the following are possible
//
//	 type MySuiteFixture { FX }
//
//	 func(s *MySuiteFixture) Test_true_assertion_failing(t *gounit.T) {
//	     t.True(false)
//	 }
//
//	 type Assertion { gounit.Suite }
//
//	func(s *Assertion) For_true_fails_if_false_value_given(t *gounit.T) {
//		fx := &MySuiteFixture{}
//	    gounit.Run(fx, t.GoT())
//		// i.e. not the fixture-suite-test fails but the testing suite's
//		// can investigate if the test would have failed.
//	    if fx.Logs == "" {
//	        t.Error("expected true assertion to fail")
//	    }
//	}
//
//	 or
//
//	 type MySuiteFixture { FX }
//
//	 func(s *MySuiteFixture) Test_true_assertion_passing(t *gounit.T) {
//	     t.True(true)
//	 }
//
//	 type Assertion { gounit.Suite }
//
//	 func(s *Assertion) For_true_passes_if_true_value_given(t *gounit.T) {
//	     fx := &MySuiteFixture{}
//	     gounit.Run(fx, t.GoT())
//	     if fx.Logs != "" {
//			// i.e. not the fixture-suite-test fails but the testing suite's
//			// test fails.
//			t.Error(fx.Logs)
//		}
//	}
type FX struct {
	FixtureLog
	gounit.Suite
	t *gounit.T
}

func (s *FX) SetUp(t *gounit.T) { s.t = t }

func (s *FX) error(args ...interface{}) {
	s.t.Log(args...)
}

func (s *FX) Error() func(args ...interface{}) {
	return s.error
}

// FixtureLog provides the general logging facility for test suites
// fixtures by implementing gounit.SuiteLogger.  A FixtureLog mustn't
// been copied once it has been used.
type FixtureLog struct {
	Logs  string
	mutex sync.Mutex
}

// log logs concurrency save given arguments to the *Logs* property.
func (fl *FixtureLog) log(args ...interface{}) {
	fl.mutex.Lock()
	defer fl.mutex.Unlock()
	fl.Logs += fmt.Sprint(args...)
}

// Logger implements the Logger interface, i.e. the suite-tests runner
// will use the returned function to implement gounit.T.Log/Logf.
func (fl *FixtureLog) Logger() func(args ...interface{}) {
	return fl.log
}

var file = func() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("fix: suites: can't determine file")
	}
	return f
}()

// TestAllSuiteTestsAreRun is a suite fixture to verify that the
// suite-test runner executes public suite-methods as tests.
type TestAllSuiteTestsAreRun struct {
	gounit.Suite
	FixtureLog
	// Exp is logged iff *A_test*-method is called
	Exp string
}

// A_test as a public method should be run by the suite-tests runner,
// i.e. log the content of *Exp*.
func (s *TestAllSuiteTestsAreRun) A_test(t *gounit.T) { t.Log(s.Exp) }

// private can't be run.
func (s *TestAllSuiteTestsAreRun) private(t *gounit.T) { t.Log("failed") }

func (fl *TestAllSuiteTestsAreRun) File() string { return file }

// TestSuiteLogging tests if a implemented SuiteLogger of a test-suite
// is used for logging.
type TestSuiteLogging struct {
	FixtureLog
	gounit.Suite
	// Exp is logged iff *Log_test*-is called
	Exp string
	// ExpFmt is logged if *Log_fmt_test*-is called
	ExpFmt string
}

// Log_test logs *Exp*.
func (s *TestSuiteLogging) Log_test(t *gounit.T) { t.Log(s.Exp) }

// Log_fmt_test logs *ExpFmt*.
func (s TestSuiteLogging) Log_fmt_test(t *gounit.T) {
	t.Logf("%s", s.ExpFmt)
}

func (fl *TestSuiteLogging) File() string { return file }

// TestSetup  has its *SetUp*-method called before each test iff it logs
// "-11-22" or "-22-11" or "-1-212" or "-1-221" or "-2-121" or "-2-112".
// NOTE this suite's tests run in parallel making an effort to randomly
// pause a setup or test execution to have different log-values for
// different test-runs.
type TestSetup struct {
	FixtureLog
	gounit.Suite
	idx uint32
	fx  gounit.Fixtures
}

func (s *TestSetup) SetUp(t *gounit.T) {
	t.Parallel()
	s.fx.Set(t, int(atomic.AddUint32(&s.idx, 1)))
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(-1 * s.fx.Int(t))
}

func (s *TestSetup) Test_A(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(s.fx.Int(t))
}

func (s *TestSetup) Test_B(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(s.fx.Int(t))
}

func (s *TestSetup) File() string {
	return file
}

// TestTearDown  has its *SetUp*-method called before each test iff it logs
// "1-12-2" or "2-21-1" or "12-1-2" or "12-2-1" or "21-2-1" or "21-1-2".
// NOTE this suite's tests run in parallel making an effort to randomly
// pause a tear-down or test execution to have different log-values for
// different test-runs.
type TestTearDown struct {
	FixtureLog
	gounit.Suite
	idx uint32
	fx  gounit.Fixtures
}

func (s *TestTearDown) SetUp(t *gounit.T) {
	t.Parallel()
	s.fx.Set(t, int(atomic.AddUint32(&s.idx, 1)))
}
func (s *TestTearDown) TearDown(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(-1 * s.fx.Del(t).(int))
}

func (s *TestTearDown) Test_A(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(s.fx.Int(t))
}

func (s *TestTearDown) Test_B(t *gounit.T) {
	if time.Now().UnixMicro()%2 == 0 {
		time.Sleep(1 * time.Millisecond)
	}
	t.Log(s.fx.Int(t))
}

func (s *TestTearDown) File() string { return file }

// TestTearDownAfterCancel implements for each possible test
// cancellation --- FailNow, FatalIfNot, FatalOn, Fatal, Fatalf --- a
// suite test while tear-down simply logs the number of called teared
// down.  gounit.T's default cancellation is overwritten by this
// test-suite suppressing the actual cancellation which has the
// consequence that tear-down is called twice once during the
// cancellation process and once after the suite-test since its
// cancellation is suppressed.  I.e. every second tear down call is
// ignored to get the expected log of "12345".
type TestTearDownAfterCancel struct {
	FixtureLog
	gounit.Suite
	idx    uint32
	logged bool
}

func (s *TestTearDownAfterCancel) TearDown(t *gounit.T) {
	if s.logged {
		// since the fatales are suppressed ignore second log call
		s.logged = false
		return
	}
	s.logged = true
	t.Log(atomic.AddUint32(&s.idx, 1))
}

func (s *TestTearDownAfterCancel) Fail_now_test(t *gounit.T) {
	t.FailNow() // should log 0
}

func (s *TestTearDownAfterCancel) Fatal_if_not_test(t *gounit.T) {
	t.FatalIfNot(false) // should log 1
}

func (s *TestTearDownAfterCancel) Fatal_on_test(t *gounit.T) {
	t.FatalOn(errors.New("")) // should log 2
}

func (s *TestTearDownAfterCancel) Fatal_test(t *gounit.T) {
	t.Fatal("") // should log 3
}

func (s *TestTearDownAfterCancel) Fatalf_test(t *gounit.T) {
	t.Fatalf("%s", "") // should log 4
}

func (s *TestTearDownAfterCancel) Cancel() func() {
	return func() {}
}

func (s *TestTearDownAfterCancel) File() string { return file }

// TestInit logs for each setup-call '-1', each tear-down-call '-2' and
// of the two tests their respective index.  Finally the Init method
// logs *InitPrefix*, i.e. log length should be 10+len(InitPrefix) and
// should start with InitPrefix iff Init was called first and all
// methods are executed as expected.
type TestInit struct {
	FixtureLog
	gounit.Suite
}

func (s *TestInit) Init(t *gounit.I) { t.Log("") }

func (s *TestInit) SetUp(t *gounit.T) {
	t.Parallel()
	t.Log(-1)
}
func (s *TestInit) TearDown(t *gounit.T) { t.Log(-2) }

func (s *TestInit) Test_a(t *gounit.T) { t.Log(0) }
func (s *TestInit) Test_b(t *gounit.T) { t.Log(1) }

func (s *TestInit) File() string { return file }

// TestFinalize logs for each setup-call '-1', each tear-down-call '-2'
// and of the two tests their respective index.  Finally the Finalize
// method logs *FinalPrefix*, i.e. log length should be
// 10+len(FinalPrefix) and should end with FinalPrefix iff Finalize was
// called last and all methods are executed as expected.
type TestFinalize struct {
	FixtureLog
	gounit.Suite
}

func (s *TestFinalize) SetUp(t *gounit.T) {
	t.Parallel()
	t.Log(-1)
}
func (s *TestFinalize) TearDown(t *gounit.T) { t.Log(-2) }

func (s *TestFinalize) Test_a(t *gounit.T) { t.Log(0) }
func (s *TestFinalize) Test_b(t *gounit.T) { t.Log(1) }

func (s *TestFinalize) Finalize(t *gounit.F) { t.Log("") }

func (s *TestFinalize) File() string { return file }

type TestCancelerImplementation struct {
	gounit.Suite
	fatalIfNot bool
	Got        map[int]bool
}

func (s *TestCancelerImplementation) log(args ...interface{}) {
	if s.Got == nil {
		s.Got = map[int]bool{}
	}
	ID := -1
	if len(args) == 0 {
		if s.Got[T_FATAL_IF_NOT] {
			panic("expected at least one argument")
		}
		s.Got[T_FATAL_IF_NOT] = true
		return
	}
	switch value := args[len(args)-1].(type) {
	case int:
		ID = value
	case string:
		id, err := strconv.Atoi(value)
		if err != nil {
			panic("expected cancellation ID; got %v")
		}
		ID = id
	}
	if ID < 0 || ID > F_FATAL_ON {
		panic(fmt.Sprintf(
			"expected cancellation ID in {%d, ...,%d}; got %d",
			T_FATAL_IF_NOT, F_FATAL_ON, ID))
	}
	s.Got[ID] = true
}

func (s *TestCancelerImplementation) Logger() func(...interface{}) {
	return s.log
}

func (s *TestCancelerImplementation) Cancel() func() {
	return func() {
		if !s.fatalIfNot {
			return
		}
		s.fatalIfNot = false
		s.log()
	}
}

func (s *TestCancelerImplementation) Init(t *gounit.I) {
	t.Fatal(I_FATAL)
	t.Fatalf("%d", I_FATALF)
	t.FatalOn(errors.New(strconv.Itoa(I_FATAL_ON)))
}

func (s *TestCancelerImplementation) Test(t *gounit.T) {
	s.fatalIfNot = true
	t.FatalIfNot(false)
	t.FatalOn(errors.New(strconv.Itoa(T_FATAL_ON)))
	t.Fatal(T_FATAL)
	t.Fatalf("%d", T_FATALF)
}

func (s *TestCancelerImplementation) Finalize(t *gounit.F) {
	t.Fatal(F_FATAL)
	t.Fatalf("%d", F_FATALF)
	t.FatalOn(errors.New(strconv.Itoa(F_FATAL_ON)))
}

func (s *TestCancelerImplementation) File() string { return file }

type TestInitFinalHaveRunTest struct {
	FixtureLog
	gounit.Suite

	// InitLog is logged by the Init-method
	InitLog string

	// FinalLog is logged by the Finalize-method
	FinalLog string

	RunT *testing.T

	Fatal string
}

func (s *TestInitFinalHaveRunTest) Init(t *gounit.I) {
	if s.RunT != t.GoT() {
		s.Fatal = "init: test has not run-test"
		t.GoT().FailNow()
	}
	t.Log(s.InitLog)
}

func (s *TestInitFinalHaveRunTest) Finalize(t *gounit.F) {
	if s.RunT != t.GoT() {
		s.Fatal = "finalize: test has not run-test"
		t.GoT().FailNow()
	}
	t.Log(s.FinalLog)
}
