package pkg2

import "testing"

// no suite but regular go tests, failing tests, with sub tests and with
// failing sub tests.

func TestPkg2_1(t *testing.T) {}

func TestPkg2_2(t *testing.T) { t.Error("") }

func TestPkg2_3(t *testing.T) {}

func TestPkg2_4(t *testing.T) {
	t.Run("pkg2_4_sub_1", func(t *testing.T) {})
	t.Run("pkg2_4_sub_2", func(t *testing.T) {})
	t.Run("pkg2_4_sub_3", func(t *testing.T) {})
	t.Run("pkg2_4_sub_4", func(t *testing.T) {})
}

func TestPkg2_5(t *testing.T) {
	t.Run("pkg2_5_sub_1", func(t *testing.T) {})
	t.Run("pkg2_5_sub_2", func(t *testing.T) { t.Error("") })
	t.Run("pkg2_5_sub_3", func(t *testing.T) {})
	t.Run("pkg2_5_sub_4", func(t *testing.T) { t.Error("") })
}
