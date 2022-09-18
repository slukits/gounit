// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gdamore/tcell/v2"
	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type ANewViewDisplaysInitiallyGiven struct{ Suite }

func (s *ANewViewDisplaysInitiallyGiven) SetUp(t *T) { t.Parallel() }

func (s *ANewViewDisplaysInitiallyGiven) Message(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}), 1)
	ee.Listen()
	t.Contains(tt.LastScreen.String(), fxMsg)
}

func (s *ANewViewDisplaysInitiallyGiven) Status(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}), 1)
	ee.Listen()
	t.Contains(tt.LastScreen.String(), fxStatus)
}

func (s *ANewViewDisplaysInitiallyGiven) Main_info(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}), 1)
	ee.Listen()
	t.Contains(tt.LastScreen.String(), fxReporting)
}

func (s *ANewViewDisplaysInitiallyGiven) Buttons(t *T) {
	ee, tt := lines.Test(t.GoT(), New(&fxInit{t: t}), 1)
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
	return fx(t, s)
}

func (s *AView) Sets_its_width_to_80_if_screen_bigger(t *T) {
	ee, tt, fx := s.fx(t)
	tt.FireResize(120, 24)

	ee.Update(fx, nil, func(e *lines.Env) {
		t.Eq(80, fx.Dim().Width())
	})
}

func (s *AView) Is_quit_able(t *T) {
	ee, tt, fx := s.fx(t)

	ee.Update(fx, nil, func(e *lines.Env) {
		t.True(fx.FF.Has(lines.Quitable))
	})

	tt.FireRune('q')
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
	t.Not.Contains(tt.Screen().String(), fxMsg)

	fx.updateMessage("")
	t.Contains(tt.Screen().String(), fxMsg)
}

func (s *AView) Updates_statusbar_with_given_string(t *T) {
	_, tt, fx := s.fx(t)
	exp := Statuser{Str: "updated status"}

	fx.updateStatus(exp)
	t.Contains(tt.Screen().String(), exp.Str)
}

func (s *AView) Updates_statusbar_with_given_numbers(t *T) {
	_, tt, fx := s.fx(t)
	exp := Statuser{Packages: 1, Suites: 2, Tests: 5, Failed: 2}

	fx.updateStatus(exp)

	t.Contains(tt.Screen().String(), fmt.Sprintf(dfltStatus, 1, 2, 5, 2))
}

func (s *AView) Status_has_green_background_if_not_failing(t *T) {
	_, tt, fx := s.fx(t)
	vw := Testing{t, tt, fx.view}
	sb := vw.StatusBar().TrimVertical()
	t.Eq(1, len(sb))

	styles, green := sb[0].Styles(), tcell.ColorGreen
	for i := range sb[0] {
		t.True(styles.Of(i).HasBG(green))
	}
}

func (s *AView) Status_has_red_background_if_failing(t *T) {
	_, tt, fx := s.fx(t)
	fx.updateStatus(
		Statuser{Packages: 1, Suites: 2, Tests: 5, Failed: 2})
	vw := Testing{t, tt, fx.view}
	sb := vw.StatusBar().TrimVertical()
	t.Eq(1, len(sb))

	styles, red, white := sb[0].Styles(), tcell.ColorRed, tcell.ColorWhite
	for i := range sb[0] {
		t.True(styles.Of(i).HasBG(red))
		t.True(styles.Of(i).HasFG(white))
	}
}

type fxFailButtonInitLabels struct {
	fxInit
	newBB         []ButtonDef
	buttonInitErr func(error)
}

func (fx *fxFailButtonInitLabels) Buttons(upd func(Buttoner)) Buttoner {
	fx.updButton = upd
	return &buttonerFX{err: fx.buttonInitErr, newBB: fx.newBB}
}

func (s *AView) Fails_buttons_init_if_ambiguous_labels(t *T) {
	fx := &fxFailButtonInitLabels{
		fxInit: fxInit{t: t},
		newBB:  []ButtonDef{{Label: "b"}, {Label: "b"}},
		buttonInitErr: func(err error) {
			t.ErrIs(err, ErrButtonLabelAmbiguity)
		},
	}
	lines.Test(t.GoT(), New(fx), 1)
}

type fxFailButtonInitRunes struct {
	fxInit
	newBB         []ButtonDef
	buttonInitErr func(error)
}

func (fx *fxFailButtonInitRunes) Buttons(upd func(Buttoner)) Buttoner {
	fx.updButton = upd
	return &buttonerFX{err: fx.buttonInitErr, newBB: fx.newBB}
}

func (s *AView) Fails_buttons_init_if_ambiguous_event_runes(t *T) {
	fx := &fxFailButtonInitLabels{
		fxInit: fxInit{t: t},
		newBB: []ButtonDef{
			{Label: "b1", Rune: 'r'},
			{Label: "b2", Rune: 'r'}},
		buttonInitErr: func(err error) {
			t.ErrIs(err, ErrButtonRuneAmbiguity)
		},
	}
	lines.Test(t.GoT(), New(fx), 1)
}

func (s *AView) Reports_button_clicks(t *T) {
	_, tt, fx := s.fx(t)

	fx.ClickButton(tt, fxBtt1)
	fx.ClickButton(tt, fxBtt2)
	fx.ClickButton(tt, fxBtt3)

	t.True(fx.bttOneReported)
	t.True(fx.bttTwoReported)
	t.True(fx.bttThreeReported)
}

func (s *AView) Reports_button_runes(t *T) {
	_, tt, fx := s.fx(t)

	tt.FireRune(fxRnBtt1)
	tt.FireRune(fxRnBtt2)
	tt.FireRune(fxRnBtt3)

	t.True(fx.bttOneReported)
	t.True(fx.bttTwoReported)
	t.True(fx.bttThreeReported)
}

func (s *AView) Fails_a_button_update_if_ambiguous_label_given(t *T) {
	_, _, fx := s.fx(t)
	fx.updButton(
		&buttonerFX{
			updBB: map[string]ButtonDef{fxBtt1: {Label: fxBtt2}},
			err: func(err error) {
				t.ErrIs(err, ErrButtonLabelAmbiguity)
			},
		})
}

func (s *AView) Fails_a_button_update_if_ambiguous_rune_given(t *T) {
	_, _, fx := s.fx(t)
	fx.updButton(
		&buttonerFX{
			updBB: map[string]ButtonDef{
				fxBtt1: {Label: "42", Rune: fxRnBtt2}},
			err: func(err error) {
				t.ErrIs(err, ErrButtonRuneAmbiguity)
			},
		})
}

func (s *AView) Removes_button_rune_on_zero_rune_update(t *T) {
	_, tt, fx := s.fx(t)
	tt.FireRune(fxRnBtt1)
	t.True(fx.bttOneReported)
	fx.bttOneReported = false

	fx.updButton(&buttonerFX{
		updBB: map[string]ButtonDef{fxBtt1: {Label: fxBtt1, Rune: 0}}})
	tt.FireRune(fxRnBtt1)

	t.Not.True(fx.bttOneReported)
}

func (s *AView) Updates_button_rune(t *T) {
	_, tt, fx := s.fx(t)

	// rune no-op for coverage
	fx.updButton(&buttonerFX{
		updBB: map[string]ButtonDef{
			fxBtt1: {Label: fxBtt1, Rune: fxRnBtt1}}})
	fx.updButton(&buttonerFX{
		updBB: map[string]ButtonDef{fxBtt1: {Label: fxBtt1Upd, Rune: 'x'}}})

	t.Not.True(fx.bttOneReported)
	tt.FireRune('x')
	t.True(fx.bttOneReported)
	t.Contains(tt.Screen().String(), fxBtt1Upd)
}

func (s *AView) Replaces_its_buttons(t *T) {
	_, tt, fx := s.fx(t)
	upd1, upd2 := false, false

	fx.updButton(&buttonerFX{
		replace: true,
		newBB: []ButtonDef{
			{Label: "upd 1", Rune: '1'},
			{Label: "upd 2", Rune: '2'}},
		listener: func(s string) {
			switch s {
			case "upd 1":
				upd1 = true
			case "upd 2":
				upd2 = true
			}
		}})

	tt.FireRune('1')
	tt.FireRune('2')
	t.True(upd1)
	t.True(upd2)
}

func (s *AView) Updates_its_main_content(t *T) {
	_, tt, fx := s.fx(t)
	exp := "first\n\nthird\nforth"

	fx.updateReporting(&reporterFX{content: ""}) // no-op, coverage
	fx.updateReporting(&reporterFX{content: exp})

	t.SpaceMatched(tt.Screen().String(), "first", "third", "forth")
}

func (s *AView) Clears_unused_main_lines(t *T) {
	_, tt, fx := s.fx(t)
	exp := "first line\nsecond\nthird\nforth\nfifth"
	fx.updateReporting(&reporterFX{content: exp})
	t.SpaceMatched(tt.Screen().String(), strings.Split(exp, "\n")...)

	exp = "\n\n2nd\n3rd\n4th"
	fx.updateReporting(&reporterFX{content: exp, flags: RpClearing})

	scr := tt.Screen().String()
	t.SpaceMatched(scr, strings.Split(exp, "\n")...)
	t.Not.True(strings.Contains(scr, "first line"))
	t.Not.True(strings.Contains(scr, "fifth"))
}

func (s *AView) Reports_click_in_reporting_component(t *T) {
	_, tt, fx := s.fx(t)

	cnt, exp := "first\n\nthird\nforth", 2
	fx.updateReporting(&reporterFX{content: cnt})

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
