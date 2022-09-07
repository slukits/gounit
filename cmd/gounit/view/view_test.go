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

type ANewViewDisplaysInitiallyGiven struct{ Suite }

func (s *ANewViewDisplaysInitiallyGiven) SetUp(t *T) { t.Parallel() }

func (s *ANewViewDisplaysInitiallyGiven) Message(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}))
	ee.Listen()
	t.Contains(tt.LastScreen.String(), fxMsg)
}

func (s *ANewViewDisplaysInitiallyGiven) Status(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}))
	ee.Listen()
	t.Contains(tt.LastScreen.String(), fxStatus)
}

func (s *ANewViewDisplaysInitiallyGiven) Main_info(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}))
	ee.Listen()
	t.Contains(tt.LastScreen.String(), fxReporting)
}

func (s *ANewViewDisplaysInitiallyGiven) Buttons(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}))
	ee.Listen()
	t.Contains(tt.LastScreen.String(), fxBtt1)
	t.Contains(tt.LastScreen.String(), fxBtt2)
}

func TestANewView(t *testing.T) {
	t.Parallel()
	Run(&ANewViewDisplaysInitiallyGiven{}, t)
}

type AView struct {
	Suite
	Fixtures
}

func (s *AView) SetUp(t *T) { t.Parallel() }

func (s *AView) TearDown(t *T) {
	quit := s.Get(t)
	if quit == nil {
		return
	}
	quit.(func())()
}

// fx creates a new view fixture (see newFX) and initializes with it a
// lines.Test whose returned Events and Testing instances are returned
// along with the view fixture.  A so obtained Events instance listens
// for ever and is quit by TearDown.
func (s *AView) fx(t *T) (*lines.Events, *lines.Testing, *viewFX) {
	fx := newFX(t)
	ee, tt := lines.Test(t.GoT(), fx, 0)
	ee.Listen()
	s.Set(t, ee.QuitListening)
	return ee, tt, fx
}

func (s *AView) Focuses_the_reporting_component(t *T) {
	ee, _, fx := s.fx(t)

	ee.Update(fx, nil, func(e *lines.Env) {
		t.Eq(fx.Report, e.Focused())
	})
}

func (s *AView) Reporting_component_is_scrollable(t *T) {
	ee, _, fx := s.fx(t)

	ee.Update(fx.Report, nil, func(e *lines.Env) {
		t.True(fx.Report.FF.Has(lines.Scrollable))
	})
}

func (s *AView) Updates_message_bar_with_given_message(t *T) {
	_, tt, fx := s.fx(t)
	exp := "updated message"

	fx.updateMessage(exp)

	t.Contains(tt.Screen().String(), exp)
}

func (s *AView) Resets_message_bar_to_default_if_zero_update(t *T) {
	_, tt, fx := s.fx(t)
	exp := "updated message"

	fx.updateMessage(exp)
	t.False(strings.Contains(tt.Screen().String(), fxMsg))

	fx.updateMessage("")
	t.Contains(tt.Screen().String(), fxMsg)
}

func (s *AView) Updates_statusbar_with_given_message(t *T) {
	_, tt, fx := s.fx(t)
	exp := "updated status"

	fx.updateStatus(exp)
	t.Contains(tt.Screen().String(), exp)
}

func (s *AView) Resets_statusbar_to_default_if_zero_update(t *T) {
	_, tt, fx := s.fx(t)
	exp := "updated status"

	fx.updateStatus(exp)
	t.False(strings.Contains(tt.Screen().String(), fxStatus))

	fx.updateStatus("")
	t.Contains(tt.Screen().String(), fxStatus)
}

type fxFailButtonInitLabels struct {
	fxInit
	buttonInitErr error
}

func (fx *fxFailButtonInitLabels) Buttons(
	upd ButtonUpd, cb func(ButtonDef) error,
) ButtonLst {
	fx.updButton = upd
	if err := cb(ButtonDef{Label: "b"}); err != nil {
		fx.t.Fatal("unexpected error: %v", err)
	}
	if err := cb(ButtonDef{Label: "b"}); err != nil {
		fx.buttonInitErr = err
	}
	return nil
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

func (fx *fxFailButtonInitRunes) Buttons(
	upd ButtonUpd, cb func(ButtonDef) error,
) ButtonLst {
	fx.updButton = upd
	if err := cb(ButtonDef{Label: "b1", Rune: '1'}); err != nil {
		fx.t.Fatal("unexpected error: %v", err)
	}
	if err := cb(ButtonDef{Label: "b2", Rune: '1'}); err != nil {
		fx.buttonInitErr = err
	}
	return nil
}

func (s *AView) Fails_buttons_init_if_ambiguous_event_runes(t *T) {
	fx := &fxFailButtonInitRunes{fxInit: fxInit{t: t}}

	lines.Test(t.GoT(), New(fx))

	t.ErrIs(fx.buttonInitErr, ErrButtonRuneAmbiguity)
}

func (s *AView) Reports_button_clicks(t *T) {
	_, tt, fx := s.fx(t)

	fx.ClickButton(tt, fxBtt1)
	fx.ClickButton(tt, fxBtt2)
	fx.ClickButton(tt, fxBtt3)

	t.True(fx.bttOneReported)
	t.True(fx.bttTwoReported)
	// not part of layout since zero label
	t.False(fx.bttThreeReported)
}

func (s *AView) Reports_button_runes(t *T) {
	_, tt, fx := s.fx(t)

	tt.FireRune(fxRnBtt1)
	tt.FireRune(fxRnBtt2)
	tt.FireRune(fxRnBtt3)

	t.True(fx.bttOneReported)
	t.True(fx.bttTwoReported)
	// not part of layout since zero label
	t.False(fx.bttThreeReported)
}

func (s *AView) Fails_a_button_update_if_ambiguous_label_given(t *T) {
	_, _, fx := s.fx(t)

	err := fx.updButton(fxBtt1, ButtonDef{Label: fxBtt2})
	t.ErrIs(err, ErrButtonLabelAmbiguity)
}

func (s *AView) Fails_a_button_update_if_ambiguous_rune_given(t *T) {
	_, _, fx := s.fx(t)

	err := fx.updButton(fxBtt1, ButtonDef{Label: "42", Rune: fxRnBtt2})
	t.ErrIs(err, ErrButtonRuneAmbiguity)
}

func (s *AView) Removes_button_rune_on_zero_rune_update(t *T) {
	_, tt, fx := s.fx(t)
	tt.FireRune(fxRnBtt1)
	t.True(fx.bttOneReported)
	fx.bttOneReported = false

	t.FatalOn(fx.updButton(fxBtt1, ButtonDef{Label: fxBtt1, Rune: 0}))
	tt.FireRune(fxRnBtt1)

	t.False(fx.bttOneReported)
}

func (s *AView) Updates_button_rune(t *T) {
	_, tt, fx := s.fx(t)

	// rune no-op for coverage
	t.FatalOn(fx.updButton(
		fxBtt1, ButtonDef{Label: fxBtt1, Rune: fxRnBtt1}))
	t.FatalOn(fx.updButton(
		fxBtt1, ButtonDef{Label: fxBtt1Upd, Rune: 'x'}))

	t.False(fx.bttOneReported)
	tt.FireRune('x')
	t.True(fx.bttOneReported)
	t.Contains(tt.Screen().String(), fxBtt1Upd)
}

func (s *AView) Updates_its_main_content(t *T) {
	_, tt, fx := s.fx(t)
	exp := "first\n\nthird\nforth"

	fx.updateReporting(&linerFX{content: ""}) // no-op, coverage
	fx.updateReporting(&linerFX{content: exp})

	t.SpaceMatched(tt.Screen().String(), "first", "third", "forth")
}

func (s *AView) Clears_unused_main_lines(t *T) {
	_, tt, fx := s.fx(t)
	exp := "first line\nsecond\nthird\nforth\nfifth"
	fx.updateReporting(&linerFX{content: exp})
	t.SpaceMatched(tt.Screen().String(), strings.Split(exp, "\n")...)

	exp = "\n\n2nd\n3rd\n4th"
	fx.updateReporting(&linerFX{content: exp, clearing: true})

	scr := tt.Screen().String()
	t.SpaceMatched(scr, strings.Split(exp, "\n")...)
	t.False(strings.Contains(scr, "first line"))
	t.False(strings.Contains(scr, "fifth"))
}

func (s *AView) Reports_click_in_reporting_component(t *T) {
	_, tt, fx := s.fx(t)

	cnt, exp := "first\n\nthird\nforth", 2
	fx.updateReporting(&linerFX{content: cnt})

	tt.FireComponentClick(fx.Report, 0, exp)
	t.Eq(exp, fx.reportedLine)
}

func (s *AView) Reporting_component_scrolls_on_context(t *T) {
	_, tt, fx := s.fx(t)

	nLines, _ := fx.twoPointFiveTimesReportedLines()
	scr := tt.ScreenOf(fx.Report).TrimHorizontal()
	t.Eq(nLines, 2*len(scr)+len(scr)/2) // fixture test

	expLine := scr[len(scr)-len(scr)/10] // scroll by 90%
	tt.FireComponentContext(fx.Report, 0, 0)
	gotScr := tt.ScreenOf(fx.Report).TrimHorizontal()

	t.Eq(expLine.String(), gotScr[0].String())
}

func (s *AView) Reporting_component_at_bottom_scrolls_to_top(t *T) {
	_, tt, fx := s.fx(t)
	_, LastLine := fx.twoPointFiveTimesReportedLines()
	firstLine := tt.ScreenOf(fx.Report).TrimHorizontal()[0].String()

	tt.FireComponentContext(fx.Report, 0, 0)
	tt.FireComponentContext(fx.Report, 0, 0)
	// fixture test: should be at bottom
	scr := tt.ScreenOf(fx.Report).TrimHorizontal()
	t.Eq(LastLine, scr[scr.Height()-1].String())

	tt.FireComponentContext(fx.Report, 0, 0)

	t.Eq(firstLine, tt.ScreenOf(fx.Report).TrimHorizontal()[0].String())
}

func TestAView(t *testing.T) {
	t.Parallel()
	Run(&AView{}, t)
}
