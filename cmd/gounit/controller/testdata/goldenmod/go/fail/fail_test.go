package gofail

import "testing"

// no suite but regular go tests with failing tests and with failing sub
// tests.
//
// This package         has     03/1  3 Go-tests/one failing
//	                    has   2/08/2  2 Go-test/8 subs/2 failing

func TestFail_1(t *testing.T) {}

func TestFail_2(t *testing.T) { t.Error("") }

func TestFail_3(t *testing.T) {}

func TestFail_4(t *testing.T) {
	t.Run("sub_4_sub_1", func(t *testing.T) {})
	t.Run("sub_4_sub_2", func(t *testing.T) {})
	t.Run("sub_4_sub_3", func(t *testing.T) {})
	t.Run("sub_4_sub_4", func(t *testing.T) {})
}

func TestFail_5(t *testing.T) {
	t.Run("sub_5_sub_1", func(t *testing.T) {})
	t.Run("sub_5_sub_2", func(t *testing.T) { t.Error("") })
	t.Run("sub_5_sub_3", func(t *testing.T) {})
	t.Run("sub_5_sub_4", func(t *testing.T) { t.Error("") })
}
