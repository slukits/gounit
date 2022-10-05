// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"testing"

	. "github.com/slukits/gounit"
)

type srcStats struct {
	Suite
	Fixtures
}

type fixtureSetter interface{ Set(*T, interface{}) }

func setFx(t *T, fs fixtureSetter) {
	testData, _ := t.FS().Data()
	tmp := t.FS().Tmp()
	testData.Child("pkgfixture").Copy(tmp)
	pkgStats, ok := newTestingPackageStat(
		tmp.Child("pkgfixture").Path())
	if !ok {
		t.Fatalf("failed to obtain package stats of %s",
			tmp.Child("pkgfixture").Path())
	}
	tt, err := pkgStats.loadTestFiles()
	t.FatalOn(err)
	fs.Set(t, &TestingPackage{abs: pkgStats.abs, files: tt})
}

func (s *srcStats) SetUp(t *T) {
	t.Parallel()
	setFx(t, s)
}

func (s *srcStats) fx(t *T) *TestingPackage {
	return s.Get(t).(*TestingPackage)
}

func (s *srcStats) Report_the_number_of_source_files(t *T) {
	stt := s.fx(t).SrcStats()
	t.Eq(5, stt.Files)
}

func (s *srcStats) Report_the_number_of_test_files(t *T) {
	stt := s.fx(t).SrcStats()
	t.Eq(2, stt.TestFiles)
}

func (s *srcStats) Reports_the_total_number_of_code_lines(t *T) {
	stt := s.fx(t).SrcStats()
	t.Eq(9, stt.Code)
}

func (s *srcStats) Reports_the_total_number_of_test_code_lines(t *T) {
	stt := s.fx(t).SrcStats()
	t.Eq(3, stt.TestCode)
}

func (s *srcStats) Reports_the_total_number_of_documenting_lines(t *T) {
	stt := s.fx(t).SrcStats()
	t.Eq(26, stt.Doc)
}

func TestSrcStats(t *testing.T) {
	t.Parallel()
	Run(&srcStats{}, t)
}
