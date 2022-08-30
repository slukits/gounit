package pkg1_2_3

import (
	"testing"

	. "github.com/slukits/gounit"
)

type Suite122_1 struct{ Suite }

func (s *Suite122_1) TestSuite122_1_1(t *T) {}
func (s *Suite122_1) TestSuite122_1_2(t *T) {}
func (s *Suite122_1) TestSuite122_1_3(t *T) {}
func (s *Suite122_1) TestSuite122_1_4(t *T) {}
func (s *Suite122_1) TestSuite122_1_5(t *T) {}

func TestSuite122_1(t *testing.T) { Run(&Suite122_1{}, t) }

type Suite122_2 struct{ Suite }

func (s *Suite122_1) TestSuite122_2_1(t *T) {}
func (s *Suite122_1) TestSuite122_2_2(t *T) {}
func (s *Suite122_1) TestSuite122_2_3(t *T) {}
func (s *Suite122_1) TestSuite122_2_4(t *T) {}
func (s *Suite122_1) TestSuite122_2_5(t *T) {}

func TestSuite122_2(t *testing.T) { Run(&Suite122_2{}, t) }
