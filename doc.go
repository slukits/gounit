// Package gounit augments the go testing framework with features aiding
// its user on matter of:
//   - naming
//   - focus
//   - documenting/specifying
//   - managing complexity
//   - maintenance
//
// To do so gounit is tailored to a very specific coding style.  In
// short
//
//	Find the most simple yet not trivial behavior of your software system
//	whose implementation will bring you closer to the overall goal and you
//	are absolutely sure you will need it.
//
//	Write a test for this behavior in the way you wish the production API
//	to expresses this behavior.
//
//	Make the test work as fast as possible by committing any sin which
//	makes it easier.
//
//	Once the test passes refactor the code until all sins are removed.
//
// The closer you are at this style the more gounit can do for you.
//
// gounit is a combination of a command cmd/gounit and a few types of
// the gounit package.  Only together they unlock gounit's full
// potential.  The command watches the module you are working on and
// gives you permanent redundancy free feedback about your progress
// through reporting test results, an outline of your project
// documentation, and – following above coding style – an outline of
// your thought processes in ways go test and go doc do not.  From the
// gounit package you will mainly use the types [gounit.Suite] and
// [gounit.T] ([gounit.S] for Init and Finalize) as well as the function
// [gounit.Run]:
//
//	import github.com/slukits/gounit
//
//	type TestedSubject struct{ gounit.Suite }
//
//	func (s *TestedSubject) Should_have_tested_behavior(t *gounit.T) {
//	    // test implementation
//	}
//
//	func TestTestedSubject(t *testing.T) {
//	    gounit.Run(&MySuite{}, t)
//	}
//
// If suites should run concurrently:
//
//	func TestTestedSubject(t *testing.T) {
//	    t.Parallel()
//	    gounit.Run(&MySuite{}, t)
//	}
//
// If all tests of a suite should run concurrently as well:
//
//	func (s *TestedSubject) SetUp(t *gounit.T) {
//	    t.Parallel()
//	}
//
// A suit test is a method of a gounit.Suite-embedder which is public,
// not special, and has exactly one argument (which then must be of type
// *testing.T but this is not validated, i.e. gounit will produce a
// panic if not).  Special methods are Init, SetUp, TearDown, Finalize.
// These methods behave as you expect: Init and Finalize are executed
// before respectively after any other method.  SetUp and TearDown are
// executed before respectively after each test.  The special methods
// along with compact assertions provided by gounit.T allow you in a
// systematic way to remove noise from your test implementations with
// the goal to make your suite-test implementation the specification of
// your production API.  NOTE:
//
//	func (s *TestedSubject) Init(t *gounit.S) {
//	    // initialize your fixture environment
//	}
//
//	func (s *TestedSubject) Should_have_tested_behavior(t *gounit.T) {
//	    // test implementation
//	}
//
//	func (s *TestedSubject) Finalize(t *gounit.S) {
//	    // tear down your fixture environment
//	}
//
// Init and Finalize have not *gounit.T but *gounit.S as argument type.
// The reason is that the argument of Init and Finalize has a different
// semantic than the argument of suite tests.  S and and T wrap
// testing.T instances of the go testing framework.  S wraps the test
// runner's testing.T instance, i.e. in above example it is
// TestTestedSubject's testing.T instance.  While T wraps a testing.T
// instance of a test runner's sub-test created to execute the suite
// test.
//
// How does gounit now fullfil the above promises?  To a great deal
// through the gounit command and how it reports test runs.  First of
// all it does them automatically if a package's source code file has
// changed.  It tries to always report back the test suite you are
// working on.  It reports the suite tests in the order they are written
// in the test file.  I.e. if I have a clumsy formulation or an
// unfitting notion in my suite test names I have it every 20 minutes or
// so in my face, i.e.  I will notice and I'll fix it.
//
// If I start with the most simple behavior I expect from an
// instance/operation and move subsequently on to more and more
// intricate behavior my test-names not only outline the documentation
// of the instance/operation but they also outline a thought process.
// Being confronted with it in short intervals keeps me focused and
// makes me discover wrong turns rather early.
//
// Since our test names express behavior instead of just repeating a
// method name the test names together with the suite and package name
// provide a semantical documentation of the production code.  The
// functional documentation is go doc's domain and given in the function
// respectively method documentation.  In summary a test-name should
// express what is expected, a method/function name/documentation should
// express what is done in detail, while git commits should focus on why
// is something done, i.e. keep it DRY.  Through special methods and
// compact assertions attached to provided [gounit.T] instance all noise
// can be systematically removed from a test implementation which lets
// the test express specification and usage of the production API.
//
// While IDE's are very feature rich when it comes to the details they
// are less so when it comes to the bird's eye view.  gounit shows you
// all your packages in the module overview, all suites of a package in
// the package overview and all tests of a suite.  While the suite tests
// are reported in the order they are written suites and packages are
// reported descending by their modification time.  That is how gounit
// helps to manage complexity.
//
// Finally when it comes to maintenance: having test names together with
// their suite and package names telling a story about the behavior of
// an certain aspect of a software system while clutter free tests
// provide the specifics about the production API makes it quite easy to
// get (back) into that certain aspect of a software system.
package gounit
