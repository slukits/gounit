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

func (s *_test) SetUp(t *T) {
	t.Parallel()
}

func (s *_test) TearDown(t *T) {
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
func (s *_test) fx(t *T, str *string) (
	*lines.Events, *lines.Testing, *viewFX, *Test,
) {
	fx := newFX(t)
	ee, tt := lines.Test(t.GoT(), fx, 0)
	ee.Listen()
	s.Set(t, ee.QuitListening)
	if str != nil {
		mckFtl(str, t)
	}
	return ee, tt, fx, &Test{t, fx.view}
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
	_, tt, _, tv := s.fx(t, &ftlMsg)

	cc := tv.CC
	tv.CC = nil
	tv.MessageBar(tt)
	t.Eq(noCmp, ftlMsg)

	tv.CC = cc[1:]
	tv.MessageBar(tt)
	t.True(strings.HasSuffix(ftlMsg, msgErr))
}

func (s *_test) Provides_message_bar_screen_portion(t *T) {
	_, tt, _, tv := s.fx(t, nil)
	t.Eq(fxMsg, tv.Trim(tv.MessageBar(tt)).String())
}

const rprErr = "expected second component to be reporting"

func (s *_test) Fails_providing_reporting_if_not_second_cmp(t *T) {
	var ftlMsg string
	_, tt, _, tv := s.fx(t, &ftlMsg)

	cc := tv.CC
	tv.CC = nil
	tv.Reporting(tt)
	t.Eq(notEnough, ftlMsg)

	tv.CC = append([]lines.Componenter{&button{}}, cc...)
	tv.Reporting(tt)
	t.True(strings.HasSuffix(ftlMsg, rprErr))
}

func (s *_test) Provides_reporting_screen_portion(t *T) {
	_, tt, _, tv := s.fx(t, nil)
	t.Eq(fxReporting, tv.Trim(tv.Reporting(tt)).String())
}

const sttErr = "expected third component to be the status bar"

func (s *_test) Fails_providing_status_bar_if_not_third_cmp(t *T) {
	var ftlMsg string
	_, tt, _, tv := s.fx(t, &ftlMsg)

	cc := tv.CC
	tv.CC = nil
	tv.StatusBar(tt)
	t.Eq(notEnough, ftlMsg)

	tv.CC = append([]lines.Componenter{&button{}}, cc...)
	tv.StatusBar(tt)
	t.True(strings.HasSuffix(ftlMsg, sttErr))
}

func (s *_test) Provides_status_bar_screen_portion(t *T) {
	_, tt, _, tv := s.fx(t, nil)
	t.Eq(fxStatus, tv.Trim(tv.StatusBar(tt)).String())
}

const bbrErr = "expected forth component to be a button bar"

func (s *_test) Fails_providing_button_bar_if_not_forth_cmp(t *T) {
	var ftlMsg string
	_, tt, _, tv := s.fx(t, &ftlMsg)

	cc := tv.CC
	tv.CC = nil
	tv.ButtonBar(tt)
	t.Eq(notEnough, ftlMsg)

	tv.CC = append([]lines.Componenter{&button{}}, cc...)
	tv.ButtonBar(tt)
	t.True(strings.HasSuffix(ftlMsg, bbrErr))
}

func (s *_test) Provides_button_bar_screen_portion(t *T) {
	_, tt, _, tv := s.fx(t, nil)
	exp := "first                                   second"
	t.Eq(exp, tv.Trim(tv.ButtonBar(tt)).String())
}

func (s *_test) Fails_button_bar_click_if_not_forth_cmp(t *T) {
	var ftlMsg string
	_, tt, _, tv := s.fx(t, &ftlMsg)

	cc := tv.CC
	tv.CC = nil
	tv.ClickButton(tt, fxBtt1)
	t.Eq(notEnough, ftlMsg)

	tv.CC = append([]lines.Componenter{&button{}}, cc...)
	tv.ClickButton(tt, fxBtt1)
	t.True(strings.HasSuffix(ftlMsg, bbrErr))

	tv.CC = cc
	tv.ClickButton(tt, fxBtt1Upd)
	t.True(strings.HasPrefix(
		ftlMsg, "gounit: view: fixture: no button labeled"))
}

func (s *_test) Clicks_requested_button(t *T) {
	_, tt, fx, tv := s.fx(t, nil)

	tv.ClickButton(tt, fxBtt2)
	t.True(fx.bttTwoReported)
}

func (s *_test) Clicks_requested_reporting_line(t *T) {
	_, tt, fx, tv := s.fx(t, nil)

	tv.ClickReporting(tt, 2)
	t.Eq(2, fx.reportedLine)
}

func TestTest(t *testing.T) {
	t.Parallel()
	Run(&_test{}, t)
}
