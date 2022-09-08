package gofailsingle

import "testing"

// no suite but regular go tests with exactly one failing test.
//
// This package         has     03/1  3 Go-tests/one failing
//	                    has   1/04/0  1 Go-test/4 subs/zero failing

func TestFailSng_1(t *testing.T) {}

func TestFailSng_2(t *testing.T) { t.Error("") }

func TestFailSng_3(t *testing.T) {}

func TestFailSng_4(t *testing.T) {
	t.Run("fs4_sub_1", func(t *testing.T) {})
	t.Run("fs4_sub_2", func(t *testing.T) {})
	t.Run("fs4_sub_3", func(t *testing.T) {})
	t.Run("fs4_sub_4", func(t *testing.T) {})
}
