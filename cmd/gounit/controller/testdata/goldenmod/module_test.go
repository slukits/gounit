/*
Package goldenmod should cover all possible scenarios for reported
testing packages of a watched Sources-instance to a controller which in
turn feeds this information in the view.

	This package         has     1/0  Go-test, zero failing
	                          1/05/0  1 suite/5 tests/0 failing
	empty sub-package    has     0/0  tests; i.e. is no package
	internal sub-pkg     has     0/0  tests; i.e. is no package
	internal/ipkg1       has  1/01/1  1 suite/1 test/1, fails compile
	internal/ipkg2       has  1/05/1  fails if vet=on; otherwise not
	internal/ipkg3       has  1/05/1  fails if race=on; otherwise not
	pkg1                 has  2/10/0  suite-tests
	pkg1/pkg1_1          has  2/10/0  suite-tests
	pkg1/pkg1_2          has     0/0  tests; i.e. is no package
	pkg1/pkg1_2/pkg1_2_1 has  2/10/0  suite-tests
	pkg1/pkg1_2/pkg1_2_2 has  2/10/0  suite-tests
	pkg1/pkg1_2/pkg1_2_3 has  2/10/0  suite-tests
	pkg2                 has     3/0  Go-tests
	                          2/08/0  Go-tests running sub-test
	pkg3                 has     3/1  Go-tests
	                          2/08/2  Go-tests running sub-test

If gounit is run in the modules root directory it should report the
summary of 10 packages 14 suites 89 tests of which 4-6 are failing
(depending if vet=on|off and race=on|off).  The initial screen would
report the packages ipkg1 and pkg2 with their failing tests and
depending on vet=on|off and race=on|off the packages ipkg2 and ipkg3
would be also reported with their respective build/test-failures.

If we would want an initial screen with the latest modified package and
the latest modified tests reported we would choose pkg1 as watched
sources reporting 5 packages 10 suites 50 tests of which zero fail.

Choosing pkg2 as watched sources let's evaluate how go tests and go
tests running sub tests are reported.

Choosing pkg3 as watched sources let's evaluate how failing go tests and
failing go tests running sub tests are reported.
*/
package goldenmod

import (
	"testing"

	. "github.com/slukits/gounit"
)

func TestModule(t *testing.T) {}

type Module struct{ Suite }

func (s *Module) Suite_test_1(t *T) {}
func (s *Module) Suite_test_2(t *T) {}
func (s *Module) Suite_test_3(t *T) {}
func (s *Module) Suite_test_4(t *T) {}
func (s *Module) Suite_test_5(t *T) {}

func TestModuleSuite(t *testing.T) {
	Run(&Module{}, t)
}
