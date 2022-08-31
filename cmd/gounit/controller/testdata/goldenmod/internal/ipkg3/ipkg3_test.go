// Package ipkg3 contains a suite with a data race error which fails a
// test and the test-command; i.e. having race turned off the suite will
// pass.
package ipkg3

import (
	"testing"

	. "github.com/slukits/gounit"
)

type SuiteI2 struct {
	Suite
	race int
}

func (s *SuiteI2) SetUp(t *T)         { t.Parallel() }
func (s *SuiteI2) TestSuiteI2_1(t *T) { s.race++ }
func (s *SuiteI2) TestSuiteI2_2(t *T) { s.race++ }
func (s *SuiteI2) TestSuiteI2_3(t *T) { s.race++ }
func (s *SuiteI2) TestSuiteI2_4(t *T) { s.race++ }
func (s *SuiteI2) TestSuiteI2_5(t *T) { s.race++ }

func TestSuiteI2_1(t *testing.T) { Run(&SuiteI2{}, t) }
