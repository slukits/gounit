package pkg1_2_3

import (
	"testing"

	. "github.com/slukits/gounit"
)

type Suite123_1 struct{ Suite }

func (s *Suite123_1) TestSuite123_1_1(t *T) {}
func (s *Suite123_1) TestSuite123_1_2(t *T) {}
func (s *Suite123_1) TestSuite123_1_3(t *T) {}
func (s *Suite123_1) TestSuite123_1_4(t *T) {}
func (s *Suite123_1) TestSuite123_1_5(t *T) {}

func TestSuite123_1(t *testing.T) { Run(&Suite123_1{}, t) }

type Suite123_2 struct{ Suite }

func (s *Suite123_1) TestSuite123_2_1(t *T) {}
func (s *Suite123_1) TestSuite123_2_2(t *T) {}
func (s *Suite123_1) TestSuite123_2_3(t *T) {}
func (s *Suite123_1) TestSuite123_2_4(t *T) {}
func (s *Suite123_1) TestSuite123_2_5(t *T) {}

func TestSuite123_2(t *testing.T) { Run(&Suite123_2{}, t) }
