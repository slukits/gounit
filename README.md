# Overview

Package gounit is an augmentation of the go testing framework for test
driven development.  It comes with a few types to implement test suites
and compact assertion to systematically remove noise from tests.  It
also comes with a command "gounit" which watches the source directory in
which it was executed given it is in the directory structure of a go
module.  The gounit command

  - lets you browse all tests of watched directory and nested packages.

  - reruns a package's tests on modification.

  - follows automatically failing tests.

  - reports test-names in a human friendly manner

  - suite-tests are reported in the order they are written.

  - provides handy switches to turn go vet, the race detector or
    source-code statistics on and off.

execute

```bash
$ go install github.com/slukits/gounit/cmd/gounit@latest
```

to install the gounit command.


![simple gounit use-case](gounit.gif)

# gounit types

From the gounit package you will mainly use the types gounit.Suite and
gounit.T (gounit.S for Init and Finalize) as well as the function
gounit.Run:


```go
	import github.com/slukits/gounit

	type TestedSubject struct{ gounit.Suite }

	func (s *TestedSubject) Should_have_tested_behavior(t *gounit.T) {
	    // test implementation
	}

	func TestTestedSubject(t *testing.T) {
	    gounit.Run(&TestedSubject{}, t)
	}

```
    
If all tests of a suite should run concurrently:
    
```go
	func (s *TestedSubject) SetUp(t *gounit.T) {
	    t.Parallel()
	}
```
    
Note that gounit also reports normal go-tests and go-tests with
sub-tests.  While on the other hand suite tests are also executed using
the "go test" command.  A suit test is a method of a
gounit.Suite-embedder which is public, not special, and has exactly one
argument (which then must be of type *gounit.T but this is not
validated, i.e. gounit will panic if not).  Special methods are Init,
SetUp, TearDown and Finalize as well as Get, Set and Del.  The first
four methods behave as you expect: Init and Finalize are executed before
respectively after all suite-tests.  SetUp and TearDown are executed
before respectively after each suite-test.  The other three methods are
considered special because they are implemented by the
gounit.Fixtures-utility and it turned out to be a quite natural use case
to embedded the Fixtures-type next to the Suite type in a test suite.
Special methods along with compact assertions provided by gounit.T allow
you in a systematic way to remove noise from your tests with the goal to
make your suite-test implementations the specification of your
production API.  While suite tests reported in the order they were
written will outline the behavior of your production code and the
thought process which led there.

NOTE:

```go
	func (s *TestedSubject) Init(t *gounit.S) {
	    // initialize your fixtures environment
	}

	func (s *TestedSubject) Should_have_tested_behavior(t *gounit.T) {
	    // test implementation
	}

	func (s *TestedSubject) Finalize(t *gounit.S) {
	    // tear down your fixtures environment
	}
```

Init and Finalize have not *gounit.T but *gounit.S as argument type.
The reason is that the argument of Init and Finalize has a different
semantic than the argument of suite tests.  S and and T wrap testing.T
instances of the go testing framework.  S wraps the suite runner's
testing.T instance, i.e. in above example it is TestTestedSubject's
testing.T instance.  While T wraps a testing.T instance of a test
runner's sub-test created to execute the suite test.  A typical
full-blown test suite (in pseudo-code) might look like this:

```go
	type testedSubject struct{
	    gounit.Suite
	    gounit.Fixtures
	    fixtureOriginal *myFixture
	}

	func (s *testedSubject) Init(t *gounit.S) {
	    s.fixtureOriginal = myInMemoryFixtureGenerator()
	}

	func (s *testedSubject) SetUp(t *gounit.T) {
	    t.Parallel()
	    s.Set(t, s.fixtureOriginal.ConcurrencySaveClone())
	}

	func (s *testedSubject) TearDown(t *gounit.T) {
	    s.Del(t).(*myFixture).CleanUp()
	}

	func (s *testedSubject) fx(t *gounit.T) *myFixture {
	    return s.Get(t).(*myFixture)
	}

	func (s *testedSubject) Has_tested_behavior(t *gounit.T) {
	    fx := s.fx(t)
	    // do something within the test specific fixated environment
	    // and assert the effect of this doing.
	}

	func (s *testedSubject) Finalize(t *gounit.S) {
	    s.fixtureOriginal.CleanUp()
	}

	func TestTestedSubject(t *testing.T) {
	    t.Parallel()
	    Run(&testedSubject{}, t)
	}
```

# gounit command

Execute gounit in the go source directory which you want it to watch.
This directory must be located inside a go module.  Click gounit's
help button to learn how its ui is working.

## Traps for the unwary

To avoid unnecessary overhead gounit investigates a go package if it is
testing.  I.e. if it at least contains one go test.  It does so by
parsing packages *_test.go files for a go test.  Lets assume you
implement a test suite and after it you implement the suite-runner which
is a go test, i.e. the package is considered testing.  If you now have a
syntax error in you suite code, i.e. the parser stops before the suite
runner is reached, then the package is considered not testing.  The
later may be confusing if you are unaware of that syntax error.

Hence before you file an issue about not reported suite tests make sure
your tests compile.

The timing which is reported to packages is the total time needed for
the system-command running a package's tests while "go test" usually
reports only the time for the tests execution.  It is rather difficult
to say something about the overhead introduced by the features of
gounit.  Timings I did suggest an overhead between 25 and 50 percent.
Practically I haven't experienced a notable difference.

Happy coding!