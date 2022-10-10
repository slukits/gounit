package gosuite

import (
	"testing"
)

func TestPass_1(t *testing.T) {}

func Test2Fail(t *testing.T) {
	t.Run("p2_sub_1", func(t *testing.T) {})
	t.Run("p2_sub_2", func(t *testing.T) { t.Errorf("failing") })
	t.Run("p2_sub_3", func(t *testing.T) {})
	t.Run("p2_sub_4", func(t *testing.T) {})
}

func Test3Pass(t *testing.T) {
	t.Run("p3_sub_1", func(t *testing.T) {})
	t.Run("p3_sub_2", func(t *testing.T) {})
	t.Run("p3_sub_3", func(t *testing.T) {})
	t.Run("p3_sub_4", func(t *testing.T) {})
}
