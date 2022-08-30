package pkg1_2_3

import (
	"testing"

	. "github.com/slukits/gounit"
)

type Suite11_1 struct{ Suite }

func (s *Suite11_1) TestSuite11_1_1(t *T) {}
func (s *Suite11_1) TestSuite11_1_2(t *T) {}
func (s *Suite11_1) TestSuite11_1_3(t *T) {}
func (s *Suite11_1) TestSuite11_1_4(t *T) {}
func (s *Suite11_1) TestSuite11_1_5(t *T) {}

func TestSuite11_1(t *testing.T) { Run(&Suite11_1{}, t) }

type Suite11_2 struct{ Suite }

func (s *Suite11_1) TestSuite11_2_1(t *T) {}
func (s *Suite11_1) TestSuite11_2_2(t *T) {}
func (s *Suite11_1) TestSuite11_2_3(t *T) {}
func (s *Suite11_1) TestSuite11_2_4(t *T) {}
func (s *Suite11_1) TestSuite11_2_5(t *T) {}

func TestSuite11_2(t *testing.T) { Run(&Suite11_2{}, t) }
