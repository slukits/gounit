package pkg2

import "testing"

// no suite but a regular go test with sub tests some of them failing.

func TestPkg2_1(t *testing.T) {
	t.Run("pkg2_1_sub_1", func(t *testing.T) {})
	t.Run("pkg2_1_sub_2", func(t *testing.T) {})
	t.Run("pkg2_1_sub_3", func(t *testing.T) {})
	t.Run("pkg2_1_sub_4", func(t *testing.T) {})
}

func TestPkg2_2(t *testing.T) {
	t.Run("pkg2_2_sub_1", func(t *testing.T) {})
	t.Run("pkg2_2_sub_2", func(t *testing.T) { t.Error("") })
	t.Run("pkg2_2_sub_3", func(t *testing.T) {})
	t.Run("pkg2_2_sub_4", func(t *testing.T) { t.Error("") })
}
