package ipkg2

import (
	"testing"

	. "github.com/slukits/gounit"
)

type SuiteI2_1 struct{ Suite }

func (s *SuiteI2_1) TestSuite11_1_1(t *T) {}
func (s *SuiteI2_1) TestSuite11_1_2(t *T) {}
func (s *SuiteI2_1) TestSuite11_1_3(t *T) {}
func (s *SuiteI2_1) TestSuite11_1_4(t *T) {}
func (s *SuiteI2_1) TestSuite11_1_5(t *T) {}

func TestSuiteI2_1(t *testing.T) { Run(&SuiteI2_1{}, t) }

type SuiteI2_2 struct{ Suite }

func (s *SuiteI2_2) TestSuite11_2_1(t *T) {}
func (s *SuiteI2_2) TestSuite11_2_2(t *T) {}
func (s *SuiteI2_2) TestSuite11_2_3(t *T) {}
func (s *SuiteI2_2) TestSuite11_2_4(t *T) {}
func (s *SuiteI2_2) TestSuite11_2_5(t *T) {}

func TestSuiteI2_2(t *testing.T) { Run(&SuiteI2_2{}, t) }
