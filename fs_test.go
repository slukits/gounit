// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"errors"
	"io/fs"
	"os"
	fp "path/filepath"
	"strings"
	"testing"
)

// tmpDir tests the behavior of a temporary directory TmpDir-instance
// td obtained from a testing instance's FS-method.
//
//	func(s *MySuite) My_suite_test(t *T) {
//	    td := t.FS().Tmp()
//	}
type tmpDir struct{ Suite }

func (s *tmpDir) SetUp(t *T) { t.Parallel() }

func (s *tmpDir) Has_reported_path_created(t *T) {
	_, err := os.Stat(t.FS().Tmp().Path())
	t.True(err == nil)
}

func (s *tmpDir) Makes_nested_directories_from_a_series_of_names(t *T) {
	exp := []string{"a", "path", "of", "directories"}

	nested, _ := t.FS().Tmp().MkTmp(exp[0], exp[1:]...)

	t.True(strings.HasSuffix(nested.Path(), fp.Join(exp...)))
	_, err := os.Stat(nested.Path())
	t.True(err == nil)
}

func (s *tmpDir) Fatales_if_directory_creation_fails(t *T) {
	exp := []string{"a", "path", "of", "directories"}
	fx, failed := NewFS(t), false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().MkdirAll(func(s string, fm fs.FileMode) error {
		return errors.New("mock-err")
	})

	fx.Tmp().MkTmp(exp[0], exp[1:]...)

	t.True(failed)
}

func (s *tmpDir) Panics_if_undoing_dir_creation_fails(t *T) {
	exp := []string{"a", "path", "of", "directories"}
	fx := NewFS(t)
	_, undo := fx.Tmp().MkTmp(exp[0], exp[1:]...)
	fx.Mock().RemoveAll(func(s string) error {
		return errors.New("mock-err")
	})

	t.Panics(func() { undo() })
}

func (s *tmpDir) Creates_file_with_given_name_and_content(t *T) {
	td := t.FS().Tmp()
	expFl, expContent := "test.txt", []byte("the answer is 42\n")

	td.MkFile(expFl, expContent)

	_, err := os.Stat(fp.Join(td.Path(), expFl))
	t.FatalIfNot(t.True(err == nil))
	gotContent, err := os.ReadFile(fp.Join(td.Path(), expFl))
	t.FatalOn(err)
	t.Eq(expContent, gotContent)
}

func (s *tmpDir) Can_undo_file_creation(t *T) {
	td := t.FS().Tmp()
	expFl, expContent := "test.txt", []byte("fearless\n")
	undo := td.MkFile(expFl, expContent)
	_, err := os.Stat(fp.Join(td.Path(), expFl))
	t.FatalIfNot(t.True(err == nil))

	undo()

	_, err = os.Stat(fp.Join(td.Path(), expFl))
	t.ErrIs(err, os.ErrNotExist)

}

func (s *tmpDir) Fatales_if_file_to_create_exists(t *T) {
	td := t.FS().Tmp()
	expFl, expContent, failed := "test.txt", []byte("fearless\n"), false

	td.MkFile(expFl, expContent)
	t.Mock().Canceler(func() { failed = true })
	td.MkFile(expFl, expContent)

	t.True(failed)
}

func (s *tmpDir) Fatales_if_file_write_fails(t *T) {
	fx := NewFS(t)
	expFl, expContent, failed := "test.txt", []byte("fearless\n"), false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().WriteFile(
		func(s string, b []byte, fm fs.FileMode) error {
			return errors.New("mock-err")
		})

	fx.Tmp().MkFile(expFl, expContent)

	t.True(failed)
}

func (s *tmpDir) Panics_if_undoing_file_creation_fails(t *T) {
	fx := NewFS(t)
	expFl, expContent := "test.txt", []byte("fearless\n")
	undo := fx.Tmp().MkFile(expFl, expContent)
	fx.Mock().Remove(func(s string) error {
		return errors.New("mock-err")
	})

	t.Panics(func() { undo() })
}

func fxFile(t *T) (td *TmpDir, name string, content []byte) {
	name, content = "test.txt", []byte("joyful\n")
	td = t.FS().Tmp()
	td.MkFile(name, content)
	return
}

func (s *tmpDir) Copies_file_to_other_dir(t *T) {
	td, expFl, expContent := fxFile(t)
	other, _ := td.Mk("other")

	td.CopyFl(expFl, other)

	t.Eq(expContent, td.FileContent(expFl))
	t.Eq(td.FileContent(expFl), other.FileContent(expFl))
}

func (s *tmpDir) Can_undo_file_copy(t *T) {}

func TestTmpDir(t *testing.T) {
	t.Parallel()
	Run(&tmpDir{}, t)
}

type Testdata struct{ Suite }

func (s *Testdata) Directory_is_created_if_not_existing(t *T) {

}

func (s *Testdata) Creation_can_be_undone(t *T) {

}

func TestTestdata(t *testing.T) {
	Run(&Testdata{}, t)
}
