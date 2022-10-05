// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package model

import (
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// SrcStats provides information about a testing package in terms of the
// number of code files, tests files, code lines, test code lines and
// documentation lines.
type SrcStats struct {

	// Files is the number of *.go files of a testing package.
	Files int

	// TestFiles is the number of *_test.go files of a testing package.
	TestFiles int

	// Code is the number of code lines of a testing package.
	Code int

	// TestCode is the number of test code lines of a testing package.
	TestCode int

	// Doc is the number of documenting lines of a testing package.
	Doc int
}

func newSrcStats(tp *TestingPackage) *SrcStats {
	stt := SrcStats{}

	ee, err := os.ReadDir(tp.abs)
	if err != nil {
		return &stt
	}
	for _, e := range ee {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".go") {
			continue
		}
		if strings.HasSuffix(e.Name(), "_test.go") {
			continue
		}
		bb, err := os.ReadFile(filepath.Join(tp.abs, e.Name()))
		if err != nil {
			continue
		}
		code, doc := linesCount(bb)
		stt.Code += code
		stt.Doc += doc
		stt.Files++
	}
	for _, t := range tp.files {
		code, doc := linesCount(t.content)
		stt.Code += code
		stt.Doc += doc
		stt.TestCode += code
	}
	stt.Files += len(tp.files)
	stt.TestFiles = len(tp.files)
	return &stt
}

var (
	ignoreCode = regexp.MustCompile(
		`^package.*$` + // package line
			`|^[\s]*$` + // empty line
			`|^[\s]*[)}\]]+[\s]*[/]*.*`, // close line
	)
	ignoreComment = regexp.MustCompile(
		`//[\s]*$` + // empty comment
			`|^[\s]*$` + // empty line
			`|^[\s]*/\*[\s]*$` + // empty start block comment
			`|^[\s]*\*/[\s]*$`, // empty end block comment
	)
	commentLine       = regexp.MustCompile(`^[\s]*//.*$`)
	startBlockComment = regexp.MustCompile(`^[\s]*/\*.*`)
	endBlockComment   = regexp.MustCompile(`^.*\*/$`)
	importLine        = regexp.MustCompile(`^[\s]*import[\s]+[^(]+$`)
	startImport       = regexp.MustCompile(`^[\s]*import[\s]+[(].*$`)
	endImport         = regexp.MustCompile(`^.*[)].*$`)
)

func commentDetector() func(l []byte) bool {

	inCommentBlock := false
	return func(l []byte) bool {
		if inCommentBlock {
			if endBlockComment.Match(l) {
				inCommentBlock = false
			}
			return true
		}
		if startBlockComment.Match(l) {
			if endBlockComment.Match(l) {
				return true
			}
			inCommentBlock = true
			return true
		}
		if commentLine.Match(l) {
			return true
		}
		return false
	}
}

func importDetector() func(l []byte) bool {
	inImport := false
	return func(l []byte) bool {
		if inImport {
			if endImport.Match(l) {
				inImport = false
			}
			return true
		}
		if startImport.Match(l) {
			if endImport.Match(l) {
				return true
			}
			inImport = true
			return true
		}
		if importLine.Match(l) {
			return true
		}
		return false
	}
}

func linesCount(dat []byte) (int, int) {

	lines := bytes.Split(dat, []byte("\n"))
	inComment, inImport := commentDetector(), importDetector()
	code, comment := 0, 0
	checkImports := true

	for _, l := range lines {
		if checkImports && inImport(l) {
			continue
		}
		if inComment(l) {
			if !ignoreComment.Match(l) {
				comment++
			}
			continue
		}
		if ignoreCode.Match(l) {
			continue
		}
		if checkImports {
			checkImports = false
		}
		code++
	}

	return code, comment
}
