// Package ipkg2 contains a suite with a vet error which expresses as a
// build error; i.e. having turned vet off the suite will pass.
package ipkg2

import (
	"fmt"
	"testing"

	. "github.com/slukits/gounit"
)

type SuiteI2 struct{ Suite }

func (s *SuiteI2) TestSuite11_1_1(t *T) {}
func (s *SuiteI2) TestSuite11_1_2(t *T) {}
func (s *SuiteI2) TestSuite11_1_3(t *T) {
	t.Log(fmt.Sprintf("vet", "error"))
}

func (s *SuiteI2) TestSuite11_1_4(t *T) {}
func (s *SuiteI2) TestSuite11_1_5(t *T) {}

func TestSuiteI2_1(t *testing.T) { Run(&SuiteI2{}, t) }
