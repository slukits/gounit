package gopass

import "testing"

// no suite but regular go tests which all pass.
// This package         has     03/0  3 Go-tests/zero failing
//	                    has   2/08/0  2 Go-test/8 subs/zero failing

func TestPass_1(t *testing.T) {}

func TestPass_2(t *testing.T) {}

func TestPass_3(t *testing.T) {}

func TestPass_4(t *testing.T) {
	t.Run("p4_sub_1", func(t *testing.T) {})
	t.Run("p4_sub_2", func(t *testing.T) {})
	t.Run("p4_sub_3", func(t *testing.T) {})
	t.Run("p4_sub_4", func(t *testing.T) {})
}

func TestPass_5(t *testing.T) {
	t.Run("p5_sub_1", func(t *testing.T) {})
	t.Run("p5_sub_2", func(t *testing.T) {})
	t.Run("p5_sub_3", func(t *testing.T) {})
	t.Run("p5_sub_4", func(t *testing.T) {})
}
