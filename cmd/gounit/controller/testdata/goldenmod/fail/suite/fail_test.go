package suite

import (
	. "github.com/slukits/gounit"
	"testing"
)

func Test1Pass(t *testing.T) {}

func Test2Pass(t *testing.T) {
	t.Run("p2_sub_1", func(t *testing.T) {})
	t.Run("p2_sub_2", func(t *testing.T) {})
	t.Run("p2_sub_3", func(t *testing.T) {})
	t.Run("p2_sub_4", func(t *testing.T) {})
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
func (s *Suite_2) Suite_test_2_4(t *T) { t.True(false) }
func (s *Suite_2) Suite_test_2_5(t *T) {}

func TestSuite_2(t *testing.T) { Run(&Suite_2{}, t) }
