package race

import (
	"testing"

	"github.com/slukits/gounit"
)

type raceSuite struct{ gounit.Suite }

func (s *raceSuite) Fails_on_race_detector(t *gounit.T) {
	RacingFunc()
}

func TestRaceSuite(t *testing.T) { gounit.Run(&raceSuite{}, t) }
