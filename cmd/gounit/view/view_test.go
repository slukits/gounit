// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package view

import (
	"fmt"
	"strings"
	"testing"

	. "github.com/slukits/gounit"
	"github.com/slukits/lines"
)

type ANewViewDisplaysInitiallyGiven struct{ Suite }

func (s *ANewViewDisplaysInitiallyGiven) SetUp(t *T) { t.Parallel() }

func (s *ANewViewDisplaysInitiallyGiven) Message(t *T) {
	tt := Fixture(t, 0, nil)
	t.Contains(tt.Screen(), fxMsg)
}

func (s *ANewViewDisplaysInitiallyGiven) Status(t *T) {
	tt := Fixture(t, 0, nil)
	t.Contains(tt.Screen(), fxStatus)
}

func (s *ANewViewDisplaysInitiallyGiven) Main_info(t *T) {
	tt := Fixture(t, 0, nil)
	t.Contains(tt.Screen(), fxReporting)
}

func (s *ANewViewDisplaysInitiallyGiven) Buttons(t *T) {
	tt := Fixture(t, 0, nil)
	t.Contains(tt.Screen(), fxBtt1)
	t.Contains(tt.Screen(), fxBtt2)
}

func TestANewView(t *testing.T) {
	t.Parallel()
	Run(&ANewViewDisplaysInitiallyGiven{}, t)
}

type AView struct {
	Suite
}

func (s *AView) SetUp(t *T) { t.Parallel() }

func (s *AView) Sets_its_width_to_80_if_screen_bigger(t *T) {
	tt := Fixture(t, 0, nil)
	tt.FireResize(120, 24)

	tt.Lines.Update(tt.Cmp, nil, func(e *lines.Env) {
		t.Eq(80, tt.Cmp.Dim().Width())
	})
}

func (s *AView) Is_quit_able(t *T) {
	tt := Fixture(t, 0, nil)

	tt.Lines.Update(tt.Cmp, nil, func(e *lines.Env) {
		t.True(tt.Cmp.FF.Has(lines.Quitable))
	})
}

func (s *AView) Updates_message_bar_with_given_message(t *T) {
	tt := Fixture(t, 0, nil)
	exp := "updated message"

	tt.UpdateMessage(exp)

	t.Contains(tt.Screen().String(), exp)
}

func (s *AView) Resets_message_bar_to_default_if_zero_update(t *T) {
	tt := Fixture(t, 0, nil)
	exp := "updated message"

	tt.UpdateMessage(exp)
	t.Not.Contains(tt.Screen(), fxMsg)

	tt.UpdateMessage("")
	t.Contains(tt.Screen(), fxMsg)
}

func (s *AView) Updates_statusbar_with_given_string(t *T) {
	tt := Fixture(t, 0, nil)
	exp := Statuser{Str: "updated status"}
	tt.UpdateStatus(exp)
	t.Contains(tt.Screen(), exp.Str)
}

func (s *AView) Updates_statusbar_with_given_numbers(t *T) {
	tt := Fixture(t, 0, nil)
	exp := Statuser{Packages: 1, Suites: 2, Tests: 5, Failed: 2}

	tt.UpdateStatus(exp)

	t.Contains(tt.Screen(), fmt.Sprintf(dfltStatus, 1, 2, 5, 2))
}

func (s *AView) Status_has_green_background_if_not_failing(t *T) {
	tt := Fixture(t, 0, nil)
	tt.UpdateStatus(Statuser{
		Packages: 1, Suites: 2, Tests: 5, Failed: 0})
	sb := tt.StatusBarCells()
	t.Eq(2, len(sb))

	l1 := sb[1]
	for i := range l1 {
		t.True(l1.HasBG(i, lines.Green))
	}
}

func (s *AView) Status_has_red_background_if_failing(t *T) {
	tt := Fixture(t, 0, nil)
	tt.UpdateStatus(Statuser{
		Packages: 1, Suites: 2, Tests: 5, Failed: 2})
	sb := tt.StatusBarCells()
	t.Eq(2, len(sb))

	l1 := sb[1]
	for i := range l1 {
		t.True(l1.HasBG(i, lines.Red))
	}
}

type fxFailButtonInitLabels struct {
	fxInit
	newBB         []ButtonDef
	buttonInitErr func(error)
}

func (fx *fxFailButtonInitLabels) Buttons(upd func(Buttoner)) Buttoner {
	return &buttonerFX{err: fx.buttonInitErr, newBB: fx.newBB}
}

func (s *AView) Fails_buttons_init_if_ambiguous_labels(t *T) {
	errReported := false
	fx := &fxFailButtonInitLabels{
		fxInit: fxInit{t: t},
		newBB:  []ButtonDef{{Label: "b"}, {Label: "b"}},
		buttonInitErr: func(err error) {
			t.ErrIs(err, ErrButtonLabelAmbiguity)
			errReported = true
		},
	}
	Fixture(t, 0, fx)
	t.True(errReported)
}

type fxFailButtonInitRunes struct {
	newBB         []ButtonDef
	buttonInitErr func(error)
}

func (fx *fxFailButtonInitRunes) Buttons(upd func(Buttoner)) Buttoner {
	return &buttonerFX{err: fx.buttonInitErr, newBB: fx.newBB}
}

func (s *AView) Fails_buttons_init_if_ambiguous_event_runes(t *T) {
	errReported := false
	fx := &fxFailButtonInitLabels{
		fxInit: fxInit{t: t},
		newBB: []ButtonDef{
			{Label: "b1", Rune: 'r'},
			{Label: "b2", Rune: 'r'}},
		buttonInitErr: func(err error) {
			t.ErrIs(err, ErrButtonRuneAmbiguity)
			errReported = true
		},
	}
	Fixture(t, 0, fx)
	t.True(errReported)
}

func (s *AView) Reports_button_clicks(t *T) {
	tt := Fixture(t, 0, nil)

	tt.ClickButton(fxBtt1)
	t.Eq(tt.ReportedButton, fxBtt1)

	tt.ClickButton(fxBtt2)
	t.Eq(tt.ReportedButton, fxBtt2)

	tt.ClickButton(fxBtt3)
	t.Eq(tt.ReportedButton, fxBtt3)
}

func (s *AView) Reports_button_runes(t *T) {
	tt := Fixture(t, 0, nil)

	tt.FireRune(fxRnBtt1)
	t.Eq(tt.ReportedButton, fxBtt1)

	tt.FireRune(fxRnBtt2)
	t.Eq(tt.ReportedButton, fxBtt2)

	tt.FireRune(fxRnBtt3)
	t.Eq(tt.ReportedButton, fxBtt3)
}

func (s *AView) Fails_a_buttons_update_if_ambiguous_label_given(t *T) {
	tt := Fixture(t, 0, nil)
	tt.UpdateButtons(
		&buttonerFX{
			updBB: map[string]ButtonDef{fxBtt1: {Label: fxBtt2}},
			err: func(err error) {
				t.ErrIs(err, ErrButtonLabelAmbiguity)
			},
		},
	)
}

func (s *AView) Fails_a_buttons_update_if_ambiguous_rune_given(t *T) {
	tt := Fixture(t, 0, nil)
	tt.UpdateButtons(
		&buttonerFX{
			updBB: map[string]ButtonDef{
				fxBtt1: {Label: "42", Rune: fxRnBtt2}},
			err: func(err error) {
				t.ErrIs(err, ErrButtonRuneAmbiguity)
			},
		})
}

func (s *AView) Removes_button_rune_on_zero_rune_update(t *T) {
	tt := Fixture(t, 0, nil)
	tt.FireRune(fxRnBtt1)
	t.Eq(tt.ReportedButton, fxBtt1)
	tt.ReportedButton = ""

	tt.UpdateButtons(&buttonerFX{
		updBB: map[string]ButtonDef{fxBtt1: {Label: fxBtt1, Rune: 0}}})
	tt.FireRune(fxRnBtt1)
	t.Eq("", tt.ReportedButton)
}

func (s *AView) Updates_button_rune(t *T) {
	tt := Fixture(t, 0, nil)

	// rune no-op for coverage
	tt.UpdateButtons(&buttonerFX{
		updBB: map[string]ButtonDef{
			fxBtt1: {Label: fxBtt1, Rune: fxRnBtt1}}})
	tt.UpdateButtons(&buttonerFX{
		updBB: map[string]ButtonDef{
			fxBtt1: {Label: fxBtt1Upd, Rune: 'x'}}})

	t.Not.True(tt.ReportedButton == fxBtt1Upd)
	tt.FireRune('x')
	t.True(tt.ReportedButton == fxBtt1Upd)
	t.Contains(tt.Screen(), fxBtt1Upd)
}

func (s *AView) Replaces_its_buttons(t *T) {
	tt := Fixture(t, 0, nil)
	upd1, upd2 := false, false

	tt.UpdateButtons(&buttonerFX{
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

func (s *AView) Updates_its_reporting_content(t *T) {
	tt := Fixture(t, 0, nil)
	exp := "first\n\nthird\nforth"

	tt.UpdateReporting(&reporterFX{content: ""}) // no-op, coverage
	tt.UpdateReporting(&reporterFX{content: exp})

	t.SpaceMatched(tt.Screen().String(), "first", "third", "forth")
}

func (s *AView) Clears_unused_main_lines(t *T) {
	tt := Fixture(t, 0, nil)
	exp := "first line\nsecond\nthird\nforth\nfifth"
	tt.UpdateReporting(&reporterFX{content: exp})
	t.SpaceMatched(tt.Screen().String(), strings.Split(exp, "\n")...)

	exp = "\n\n2nd\n3rd\n4th"
	tt.UpdateReporting(&reporterFX{content: exp, flags: RpClearing})

	scr := tt.Screen().String()
	t.SpaceMatched(scr, strings.Split(exp, "\n")...)
	t.Not.True(strings.Contains(scr, "first line"))
	t.Not.True(strings.Contains(scr, "fifth"))
}

func (s *AView) Reports_click_in_reporting_component(t *T) {
	tt := Fixture(t, 0, nil)
	tt.Lines.Update(tt.ReportCmp, nil, func(e *lines.Env) {
		fmt.Fprint(e, "first\n\nsecond\nthird")
	})

	tt.FireComponentClick(tt.ReportCmp, 0, 2)
	t.Eq(2, tt.ReportedLine)
}

func (s *AView) Reporting_component_scrolls_on_context(t *T) {
	tt := Fixture(t, 0, nil)

	nLines, _ := tt.TwoPointFiveTimesReportedLines()
	scr := tt.ScreenOf(tt.ReportCmp).Trimmed()
	t.Eq(nLines, 2*len(scr)+len(scr)/2) // fixture test

	expLine := scr[len(scr)-len(scr)/10] // scroll by 90%
	tt.FireComponentContext(tt.ReportCmp, 0, 0)
	gotScr := tt.ScreenOf(tt.ReportCmp).Trimmed()

	t.Eq(expLine, gotScr[0])
}

func (s *AView) Reporting_component_at_bottom_scrolls_to_top(t *T) {
	tt := Fixture(t, 0, nil)
	_, LastLine := tt.TwoPointFiveTimesReportedLines()
	firstLine := tt.ScreenOf(tt.ReportCmp).Trimmed()[0]

	tt.FireComponentContext(tt.ReportCmp, 0, 0)
	tt.FireComponentContext(tt.ReportCmp, 0, 0)
	// fixture test: should be at bottom
	scr := tt.ScreenOf(tt.ReportCmp).Trimmed()
	t.Eq(LastLine, scr[len(scr)-1])

	tt.FireComponentContext(tt.ReportCmp, 0, 0)
	t.Eq(firstLine, tt.ScreenOf(tt.ReportCmp).Trimmed()[0])
}

func TestAView(t *testing.T) {
	t.Parallel()
	Run(&AView{}, t)
}
