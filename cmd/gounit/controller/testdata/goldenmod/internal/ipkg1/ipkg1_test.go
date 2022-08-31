// Package ipkg1 fails to compile its test binary.
package ipkg1

import (
	"testing"

	. "github.com/slukits/gounit"
)

type SuiteIpkg1 struct{ Suite }

func (s *SuiteIpkg1) Fails_compiling(t *T) { t.True("false") }

func TestSuiteIpkg1(t *testing.T) { Run(&SuiteIpkg1{}, t) }
