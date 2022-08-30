package pkg1

import (
	"testing"

	. "github.com/slukits/gounit"
)

// suite tests; few of them failing

type Suite1 struct{ Suite }

func (s *Suite1) TestSuite1_1(t *T) {}
func (s *Suite1) TestSuite1_2(t *T) { t.True(false) }
func (s *Suite1) TestSuite1_3(t *T) {}
func (s *Suite1) TestSuite1_4(t *T) { t.True(false) }
func (s *Suite1) TestSuite1_5(t *T) {}

func TestSuite1(t *testing.T) { Run(&Suite1{}, t) }

type Suite2 struct{ Suite }

func (s *Suite1) TestSuite2_1(t *T) {}
func (s *Suite1) TestSuite2_2(t *T) {}
func (s *Suite1) TestSuite2_3(t *T) {}
func (s *Suite1) TestSuite2_4(t *T) {}
func (s *Suite1) TestSuite2_5(t *T) {}

func TestSuite2(t *testing.T) { Run(&Suite2{}, t) }
