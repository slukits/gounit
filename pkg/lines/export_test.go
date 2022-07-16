package lines

import (
	"github.com/gdamore/tcell/v2"
)

type ScreenFactoryer = screenFactoryer

// SetScreenFactory allows to mock up tcell's screen generation for
// error handling testing.  Provided factory instance must implement
// NewScreen() (tcell.Screen, error)
// NewSimulationScreen() tcell.Screen
func SetScreenFactory(f ScreenFactoryer) {
	screenFactory = f
}

func DefaultScreenFactory() ScreenFactoryer {
	return &defaultFactory{}
}

func ExtractLib(v *Screen) tcell.Screen { return v.lib }

func GetLib(rg *Events) tcell.Screen { return rg.scr.lib }
