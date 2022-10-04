package vet

import (
	"testing"

	"github.com/slukits/gounit"
)

type vetSuite struct{ gounit.Suite }

func (s *vetSuite) Fails_if_vetted(t *gounit.T) {
	t.True(SomeFunc())
}

func TestVetSuite(t *testing.T) { gounit.Run(&vetSuite{}, t) }
