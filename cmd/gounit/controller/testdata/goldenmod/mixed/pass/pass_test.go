package pass

import (
	. "github.com/slukits/gounit"
	"testing"
)

// go test and suites that pass
// This package         has     03/0  3 Go-tests/zero failing
//	                    has   2/08/0  2 Go-test/8 subs/zero failing

func TestPass_1(t *testing.T) {}

func TestPass_2(t *testing.T) {}

func TestPass_3(t *testing.T) {}

func TestPass_4(t *testing.T) {
	t.Run("p4_sub_1", func(t *testing.T) {})
	t.Run("p4_sub_2", func(t *testing.T) {})
	t.Run("p4_sub_3", func(t *testing.T) {})
	t.Run("p4_sub_4", func(t *testing.T) {})
}

func TestPass_5(t *testing.T) {
	t.Run("p5_sub_1", func(t *testing.T) {})
	t.Run("p5_sub_2", func(t *testing.T) {})
	t.Run("p5_sub_3", func(t *testing.T) {})
	t.Run("p5_sub_4", func(t *testing.T) {})
}

type Suite_1 struct{ Suite }

func (s *Suite_1) Suite_test_1_1(t *T) {}
func (s *Suite_1) Suite_test_1_2(t *T) {}
func (s *Suite_1) Suite_test_1_3(t *T) {}
func (s *Suite_1) Suite_test_1_4(t *T) {}
func (s *Suite_1) Suite_test_1_5(t *T) {}

func TestSuite_1(t *testing.T) { Run(&Suite_1{}, t) }

type Suite_2 struct{ Suite }

func (s *Suite_1) Suite_test_2_1(t *T) {}
func (s *Suite_1) Suite_test_2_2(t *T) {}
func (s *Suite_1) Suite_test_2_3(t *T) {}
func (s *Suite_1) Suite_test_2_4(t *T) {}
func (s *Suite_1) Suite_test_2_5(t *T) {}

func TestSuite_2(t *testing.T) { Run(&Suite_2{}, t) }

type Suite_3 struct{ Suite }

func (s *Suite_3) Suite_test_3_1(t *T) {}
func (s *Suite_3) Suite_test_3_2(t *T) {}
func (s *Suite_3) Suite_test_3_3(t *T) {}
func (s *Suite_3) Suite_test_3_4(t *T) {}
func (s *Suite_3) Suite_test_3_5(t *T) {}

func TestSuite_3(t *testing.T) { Run(&Suite_3{}, t) }

type Suite_4 struct{ Suite }

func (s *Suite_4) Suite_test_4_1(t *T) {}
func (s *Suite_4) Suite_test_4_2(t *T) {}
func (s *Suite_4) Suite_test_4_3(t *T) {}
func (s *Suite_4) Suite_test_4_4(t *T) {}
func (s *Suite_4) Suite_test_4_5(t *T) {}

func TestSuite_4(t *testing.T) { Run(&Suite_4{}, t) }

type Suite_5 struct{ Suite }

func (s *Suite_5) Suite_test_5_1(t *T) {}
func (s *Suite_5) Suite_test_5_2(t *T) {}
func (s *Suite_5) Suite_test_5_3(t *T) {}
func (s *Suite_5) Suite_test_5_4(t *T) {}
func (s *Suite_5) Suite_test_5_5(t *T) {}

func TestSuite_5(t *testing.T) { Run(&Suite_5{}, t) }
