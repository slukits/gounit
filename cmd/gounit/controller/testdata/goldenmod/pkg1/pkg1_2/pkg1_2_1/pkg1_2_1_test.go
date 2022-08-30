package pkg1_2_3

import (
	"testing"

	. "github.com/slukits/gounit"
)

type Suite121_1 struct{ Suite }

func (s *Suite121_1) TestSuite121_1_1(t *T) {}
func (s *Suite121_1) TestSuite121_1_2(t *T) {}
func (s *Suite121_1) TestSuite121_1_3(t *T) {}
func (s *Suite121_1) TestSuite121_1_4(t *T) {}
func (s *Suite121_1) TestSuite121_1_5(t *T) {}

func TestSuite121_1(t *testing.T) { Run(&Suite121_1{}, t) }

type Suite121_2 struct{ Suite }

func (s *Suite121_1) TestSuite121_2_1(t *T) {}
func (s *Suite121_1) TestSuite121_2_2(t *T) {}
func (s *Suite121_1) TestSuite121_2_3(t *T) {}
func (s *Suite121_1) TestSuite121_2_4(t *T) {}
func (s *Suite121_1) TestSuite121_2_5(t *T) {}

func TestSuite121_2(t *testing.T) { Run(&Suite121_2{}, t) }
