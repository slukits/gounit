package gonly

import (
	"testing"
)

func TestPass_1(t *testing.T) {}

func TestFail_2(t *testing.T) { t.Errorf("failing") }

func TestPass_3(t *testing.T) {}

func TestPass_4(t *testing.T) {
	t.Run("p4_sub_1", func(t *testing.T) {})
	t.Run("p4_sub_2", func(t *testing.T) { t.Errorf("failing") })
	t.Run("p4_sub_3", func(t *testing.T) {})
	t.Run("p4_sub_4", func(t *testing.T) {})
}

func TestPass_5(t *testing.T) {
	t.Run("p5_sub_1", func(t *testing.T) {})
	t.Run("p5_sub_2", func(t *testing.T) {})
	t.Run("p5_sub_3", func(t *testing.T) {})
	t.Run("p5_sub_4", func(t *testing.T) {})
}
