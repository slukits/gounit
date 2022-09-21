package logging

import (
	"github.com/slukits/gounit"
	"testing"
)

func TestGoLog(t *testing.T) {
	t.Log("go-test log")
}

func TestGoSuiteLog(t *testing.T) {
	t.Run("test_go_sub_log", func(t *testing.T) {
		t.Log("go sub-test log")
	})
}

type suite struct{ gounit.Suite }

func (s *suite) Init(t *gounit.S) {
	t.Log("suite init log")
}

func (s *suite) Finalize(t *gounit.S) {
	t.Log("suite finalize log")
}

func (s *suite) Suite_test_log(t *gounit.T) { t.Log("suite-test log") }

func TestSuite(t *testing.T) { gounit.Run(&suite{}, t) }
