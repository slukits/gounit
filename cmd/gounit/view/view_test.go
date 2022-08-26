// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"strings"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type ANewView struct{ Suite }

func (s *ANewView) SetUp(t *T) { t.Parallel() }

func (s *ANewView) Displays_initially_given_message(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}))
	ee.Listen()
	t.Contains(tt.LastScreen, fxMsg)
}

func (s *ANewView) Displays_initially_given_status(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}))
	ee.Listen()
	t.Contains(tt.LastScreen, fxStatus)
}

func (s *ANewView) Displays_initially_given_main_info(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}))
	ee.Listen()
	t.Contains(tt.LastScreen, fxMain)
}

func (s *ANewView) Displays_initially_given_buttons(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}))
	ee.Listen()
	t.Contains(tt.LastScreen, fxBtt1)
	t.Contains(tt.LastScreen, fxBtt2)
}

func TestANewView(t *testing.T) {
	t.Parallel()
	Run(&ANewView{}, t)
}

type AView struct{ Suite }

func (s *AView) SetUp(t *T) { t.Parallel() }

func (s *AView) Updates_message_bar_with_given_message(t *T) {
	fx, exp := &fxInit{t: t}, "updated message"
	ee, tt := lines.Test(t.GoT(), New(fx), 2)
	ee.Listen()
	fx.updateMessageBar(exp)
	t.Contains(tt.LastScreen, exp)
}

func (s *AView) Resets_message_bar_to_default_if_zero_update(t *T) {
	fx, exp := &fxInit{t: t}, "updated message"
	ee, tt := lines.Test(t.GoT(), New(fx), 3)
	ee.Listen()
	fx.updateMessageBar(exp)
	t.False(strings.Contains(tt.String(), fxMsg))
	fx.updateMessageBar("")
	t.Contains(tt.LastScreen, fxMsg)
}

func (s *AView) Updates_statusbar_with_given_message(t *T) {
	fx, exp := &fxInit{t: t}, "updated status"
	ee, tt := lines.Test(t.GoT(), New(fx), 2)
	ee.Listen()
	fx.updateStatusbar(exp)
	t.Contains(tt.LastScreen, exp)
}

func (s *AView) Resets_statusbar_to_default_if_zero_update(t *T) {
	fx, exp := &fxInit{t: t}, "updated status"
	ee, tt := lines.Test(t.GoT(), New(fx), 3)
	ee.Listen()
	fx.updateStatusbar(exp)
	t.False(strings.Contains(tt.String(), fxStatus))
	fx.updateStatusbar("")
	t.Contains(tt.LastScreen, fxStatus)
}

type fxFailButtonInitLabels struct {
	fxInit
	buttonInitErr error
}

func (fx *fxFailButtonInitLabels) ForButton(
	cb func(ButtonDef, ButtonUpdater) error,
) {
	if err := cb(ButtonDef{Label: "b"}, nil); err != nil {
		fx.t.Fatal("unexpected error: %v", err)
	}
	if err := cb(ButtonDef{Label: "b"}, nil); err != nil {
		fx.buttonInitErr = err
	}
}

func (s *AView) Fails_buttons_init_if_ambiguous_labels(t *T) {
	fx := &fxFailButtonInitLabels{fxInit: fxInit{t: t}}

	lines.Test(t.GoT(), New(fx))

	t.ErrIs(fx.buttonInitErr, ErrButtonLabelAmbiguity)
}

type fxFailButtonInitRunes struct {
	fxInit
	buttonInitErr error
}

func (fx *fxFailButtonInitRunes) ForButton(
	cb func(ButtonDef, ButtonUpdater) error,
) {
	if err := cb(ButtonDef{Label: "b1", Rune: '1'}, nil); err != nil {
		fx.t.Fatal("unexpected error: %v", err)
	}
	if err := cb(ButtonDef{Label: "b2", Rune: '1'}, nil); err != nil {
		fx.buttonInitErr = err
	}
}

func (s *AView) Fails_buttons_init_if_ambiguous_event_runes(t *T) {
	fx := &fxFailButtonInitRunes{fxInit: fxInit{t: t}}

	lines.Test(t.GoT(), New(fx))

	t.ErrIs(fx.buttonInitErr, ErrButtonRuneAmbiguity)
}

func (s *AView) Reports_button_clicks(t *T) {
	fx := newFX(t)
	_, tt := lines.Test(t.GoT(), fx, 5)

	fx.ClickButton(tt, fxBtt3) // first because not changing countdown
	fx.ClickButton(tt, fxBtt1)
	fx.ClickButton(tt, fxBtt2)

	t.True(fx.bttOneReported)
	// no listener provided
	t.False(fx.bttTwoReported)
	// not part of layout since zero label
	t.False(fx.bttThreeReported)
}

func (s *AView) Reports_button_runes(t *T) {
	fx := newFX(t)
	_, tt := lines.Test(t.GoT(), fx, 4)

	tt.FireRune(fxRnBtt1)
	tt.FireRune(fxRnBtt2)
	tt.FireRune(fxRnBtt3)

	t.True(fx.bttOneReported)
	// no listener provided
	t.False(fx.bttTwoReported)
	// not part of layout since zero label
	t.False(fx.bttThreeReported)
}

func (s *AView) Fails_a_button_update_if_ambiguous_label_given(t *T) {
	fx := &fxInit{t: t}
	ee, _ := lines.Test(t.GoT(), New(fx), 1)
	ee.Listen()
	err := fx.updBtt1(ButtonDef{Label: fxBtt2})
	t.ErrIs(err, ErrButtonLabelAmbiguity)
}

func (s *AView) Fails_a_button_update_if_ambiguous_rune_given(t *T) {
	fx := &fxInit{t: t}
	ee, _ := lines.Test(t.GoT(), New(fx), 1)
	ee.Listen()
	err := fx.updBtt1(ButtonDef{Label: "42", Rune: fxRnBtt2})
	t.ErrIs(err, ErrButtonRuneAmbiguity)
}

func (s *AView) Removes_button_rune_on_zero_rune_update(t *T) {
	fx := &fxInit{t: t}
	_, tt := lines.Test(t.GoT(), New(fx), 4)

	tt.FireRune(fxRnBtt1)
	t.True(fx.bttOneReported)
	fx.bttOneReported = false
	t.FatalOn(fx.updBtt1(ButtonDef{Label: fxBtt1, Rune: 0}))
	tt.FireRune(fxRnBtt1)

	t.False(fx.bttOneReported)
}

func (s *AView) Updates_button_rune(t *T) {
	fx := &fxInit{t: t}
	ee, tt := lines.Test(t.GoT(), New(fx), 5)
	ee.Listen()
	// rune no-op for coverage
	t.FatalOn(fx.updBtt1(ButtonDef{Label: fxBtt1, Rune: fxRnBtt1}))
	tt.FireRune(fxRnBtt2)
	t.False(fx.bttTwoReported)

	t.FatalOn(fx.updBtt2(ButtonDef{Label: "hurz", Rune: 'x',
		Listener: func(label string) {
			fx.bttTwoReported = true
		}}))
	tt.FireRune('x')

	t.True(fx.bttTwoReported)
	t.Contains(tt.LastScreen, "hurz")
}

func TestAView(t *testing.T) {
	t.Parallel()
	Run(&AView{}, t)
}
