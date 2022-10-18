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

type _test struct {
	Suite
	Fixtures
}

func (s *_test) SetUp(t *T) { t.Parallel() }

// fx creates and returns a new view fixture and in case of a given
// string pointer test-error reporting is mocked up to report to given
// string pointer.
func (s *_test) fx(t *T, str *string, i ...Initer) *Testing {
	var tt *Testing
	if len(i) == 0 {
		tt = Fixture(t, 0, nil)
	} else {
		tt = Fixture(t, 0, i[0])
	}

	if str != nil {
		mckFtl(str, t)
	}
	return tt
}

func mckFtl(msg *string, t *T) {
	t.Mock().Canceler(func() {})
	t.Mock().Logger(func(ii ...interface{}) {
		ss := []string{}
		for _, i := range ii {
			ss = append(ss, fmt.Sprintf("%v", i))
		}
		*msg = strings.Join(ss, " ")
	})
}

const (
	noCmp  = "gounit: view: fixture: no ui components"
	msgErr = "expected first component to be the message bar"
)

func (s *_test) Fails_providing_message_bar_if_not_first_cmp(t *T) {
	var ftlMsg string
	tt := s.fx(t, &ftlMsg)

	cc := tt.CC
	tt.CC = nil
	tt.MessageBarCells()
	t.Eq(noCmp, ftlMsg)

	tt.CC = cc[1:]
	tt.MessageBarCells()
	t.True(strings.HasSuffix(ftlMsg, msgErr))
}

func (s *_test) Provides_message_bar_screen_portion(t *T) {
	tt := s.fx(t, nil)
	t.Eq(fxMsg, tt.MessageBarCells().Trimmed())
}

const rprErr = "expected second component to be reporting"

func (s *_test) Fails_providing_reporting_if_not_second_cmp(t *T) {
	var ftlMsg string
	tt := s.fx(t, &ftlMsg)

	cc := tt.CC
	tt.CC = nil
	tt.ReportCells()
	t.Eq(notEnough, ftlMsg)

	tt.CC = append([]lines.Componenter{&button{}}, cc...)
	tt.ReportCells()
	t.True(strings.HasSuffix(ftlMsg, rprErr))
}

func (s *_test) Provides_reporting_screen_portion(t *T) {
	tt := s.fx(t, nil)
	t.Eq(fxReporting, tt.ReportCells().Trimmed())
}

const sttErr = "expected third component to be the status bar"

func (s *_test) Fails_providing_status_bar_if_not_third_cmp(t *T) {
	var ftlMsg string
	tt := s.fx(t, &ftlMsg)

	cc := tt.CC
	tt.CC = nil
	tt.StatusBarCells()
	t.Eq(notEnough, ftlMsg)

	tt.CC = append([]lines.Componenter{&button{}}, cc...)
	tt.StatusBarCells()
	t.True(strings.HasSuffix(ftlMsg, sttErr))
}

func (s *_test) Provides_status_bar_screen_portion(t *T) {
	tt := s.fx(t, nil)
	t.Eq(fxStatus, tt.StatusBarCells().Trimmed())
}

const bbrErr = "expected forth component to be a button bar"

func (s *_test) Fails_providing_button_bar_if_not_forth_cmp(t *T) {
	var ftlMsg string
	tt := s.fx(t, &ftlMsg)

	cc := tt.CC
	tt.CC = nil
	tt.ButtonBarCells()
	t.Eq(notEnough, ftlMsg)

	tt.CC = append([]lines.Componenter{&button{}}, cc...)
	tt.ButtonBarCells()
	t.True(strings.HasSuffix(ftlMsg, bbrErr))
}

func (s *_test) Provides_button_bar_screen_portion(t *T) {
	tt := s.fx(t, nil)
	t.SpaceMatched(tt.ButtonBarCells().Trimmed(), "first", "second")
}

func (s *_test) Fails_button_bar_click_if_not_forth_cmp(t *T) {
	var ftlMsg string
	tt := s.fx(t, &ftlMsg)

	cc := tt.CC
	tt.CC = nil
	tt.ClickButton(fxBtt1)
	t.Eq(notEnough, ftlMsg)

	tt.CC = append([]lines.Componenter{&button{}}, cc...)
	tt.ClickButton(fxBtt1)
	t.True(strings.HasSuffix(ftlMsg, bbrErr))

	tt.CC = cc
	tt.ClickButton(fxBtt1Upd)
	t.True(strings.HasPrefix(
		ftlMsg, "gounit: view: fixture: no button labeled"))
}

func (s *_test) Clicks_requested_button(t *T) {
	tt := Fixture(t, 0, nil)

	tt.ClickButton(fxBtt2)
	t.Eq(tt.ReportedButton, fxBtt2)
}

func (s *_test) Clicks_requested_reporting_line(t *T) {
	tt := Fixture(t, 0, nil)
	tt.Lines.Update(tt.ReportCmp, nil, func(e *lines.Env) {
		fmt.Fprint(e, "first\nsecond\nthird")
	})

	tt.ClickReporting(2)
	t.Eq(2, tt.ReportedLine)
}

func (s *_test) Updates_message_bar(t *T) {
	tt := Fixture(t, 0, nil)
	t.Not.Contains(tt.MessageBarCells(), "updated msg")
	tt.UpdateMessage("updated msg")
	t.True(t.Contains(tt.MessageBarCells(), "updated msg"))
}

func (s *_test) Updates_reporting_component(t *T) {
	tt, exp := Fixture(t, 0, nil), "first\nseconde\nthird"
	t.Not.SpaceMatched(tt.ReportCells(), exp)
	tt.UpdateReporting(&reporterFX{content: exp})
	t.SpaceMatched(tt.ReportCells(), exp)
}

func (s *_test) Updates_status_bar(t *T) {
	tt := Fixture(t, 0, nil)
	t.Not.Contains(tt.StatusBarCells(), "updated status")
	tt.UpdateStatus(Statuser{Str: "updated status"})
	t.True(t.Contains(tt.StatusBarCells(), "updated status"))
}

func (s *_test) Updates_buttons(t *T) {
	tt, exp := Fixture(t, 0, nil), []string{"[e]ins", "[z]wei", "[d]rei"}
	t.Not.SpaceMatched(tt.ButtonBarCells(), exp...)
	tt.UpdateButtons(&buttonerFX{
		newBB: []ButtonDef{
			{Label: "eins", Rune: 'e'},
			{Label: "zwei", Rune: 'z'},
			{Label: "drei", Rune: 'd'},
		},
		replace: true,
	})
	t.SpaceMatched(tt.ButtonBarCells(), exp...)
}

func TestTest(t *testing.T) {
	t.Parallel()
	Run(&_test{}, t)
}
