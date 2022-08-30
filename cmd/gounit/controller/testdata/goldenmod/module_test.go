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
