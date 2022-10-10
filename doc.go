/*
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

From the gounit package you will mainly use the types [gounit.Suite] and
[gounit.T] ([gounit.S] for Init and Finalize) as well as the function
[gounit.Run]:

	import github.com/slukits/gounit

	type TestedSubject struct{ gounit.Suite }

	func (s *TestedSubject) Should_have_tested_behavior(t *gounit.T) {
	    // test implementation
	}

	func TestTestedSubject(t *testing.T) {
	    gounit.Run(&TestedSubject{}, t)
	}

If all tests of a suite should run concurrently:

	func (s *TestedSubject) SetUp(t *gounit.T) {
	    t.Parallel()
	}

Note that gounit also reports normal go-tests and go-tests with
sub-tests.  While on the other hand suite tests are also executed using
the "go test" command.  A suit test is a method of a
gounit.Suite-embedder which is public, not special, and has exactly one
argument (which then must be of type *gounit.T but this is not
validated, i.e. gounit will produce a panic if not).  Special methods
are Init, SetUp, TearDown and Finalize as well as Get, Set and Del.  The
first four methods behave as you expect: Init and Finalize are executed
before respectively after all suite-tests.  SetUp and TearDown are
executed before respectively after each suite-test.  The other three
methods are considered special because they are implemented by the
[gounit.Fixtures]-utility and it turned out to be a quite natural use
case to embedded the Fixtures-type next to the Suite type in a test
suite.  Special methods along with compact assertions provided by
gounit.T allow you in a systematic way to remove noise from your tests
with the goal to make your suite-test implementations the specification
of your production API.  While suite tests reported in the order they
were written will outline the behavior of your production code and the
thought process which led there.

NOTE:

	func (s *TestedSubject) Init(t *gounit.S) {
	    // initialize your fixtures environment
	}

	func (s *TestedSubject) Should_have_tested_behavior(t *gounit.T) {
	    // test implementation
	}

	func (s *TestedSubject) Finalize(t *gounit.S) {
	    // tear down your fixtures environment
	}

Init and Finalize have not *gounit.T but *gounit.S as argument type.
The reason is that the argument of Init and Finalize has a different
semantic than the argument of suite tests.  S and and T wrap testing.T
instances of the go testing framework.  S wraps the suite runner's
testing.T instance, i.e. in above example it is TestTestedSubject's
testing.T instance.  While T wraps a testing.T instance of a test
runner's sub-test created to execute the suite test.  A typical
full-blown test suite (in pseudo-code) might look like this:

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
*/
package gounit
