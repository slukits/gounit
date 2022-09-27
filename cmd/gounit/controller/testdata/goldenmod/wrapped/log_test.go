package wrapped

import (
	"github.com/slukits/gounit"
	"testing"
)

type suite struct{ gounit.Suite }

const logTxt = "Lorem ipsum dolor sit amet, consectetur adipiscing " +
	"elit. Morbi id mi rutrum, pretium ipsum et, gravida dui. " +
	"Vestibulum et sapien et diam interdum gravida sit amet quis " +
	"leo. Suspendisse ac nisi sit amet erat eleifend bibendum. Sed " +
	"eu tincidunt arcu, sit amet pretium arcu. Nam urna eros, " +
	"aliquet sed mi vitae, consectetur consequat purus. Donec " +
	"tincidunt dictum velit, at dictum quam tincidunt ut. " +
	"Pellentesque vel dolor lacinia, dictum justo sit amet, " +
	"bibendum ex. Maecenas sit amet pellentesque leo."

func (s *suite) Suite_test_log(t *gounit.T) { t.Log(logTxt) }

func TestSuite(t *testing.T) { gounit.Run(&suite{}, t) }
