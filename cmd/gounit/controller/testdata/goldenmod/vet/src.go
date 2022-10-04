// Package vet fails if vet is turned on and passes if vet is turned
// off.
package vet

import (
	"fmt"
	"strings"
)

func SomeFunc() bool {
	return strings.Contains(fmt.Sprintf("42 %s"), "42")
}
