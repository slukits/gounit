package pass

import (
	. "github.com/slukits/gounit"
	"testing"
)

type Suite_3 struct{ Suite }

func (s *Suite_3) Suite_test_3_1(t *T) {}
func (s *Suite_3) Suite_test_3_2(t *T) {}
func (s *Suite_3) Suite_test_3_3(t *T) {}
func (s *Suite_3) Suite_test_3_4(t *T) {}
func (s *Suite_3) Suite_test_3_5(t *T) {}

func TestSuite_3(t *testing.T) { Run(&Suite_3{}, t) }
