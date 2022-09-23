package pass

import (
	. "github.com/slukits/gounit"
	"testing"
)

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

func (s *Suite_2) Suite_test_2_1(t *T) {}
func (s *Suite_2) Suite_test_2_2(t *T) {}
func (s *Suite_2) Suite_test_2_3(t *T) {}
func (s *Suite_2) Suite_test_2_4(t *T) {}
func (s *Suite_2) Suite_test_2_5(t *T) {}

func TestSuite_2(t *testing.T) { Run(&Suite_2{}, t) }
