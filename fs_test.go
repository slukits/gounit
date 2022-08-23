// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"errors"
	"io"
	"io/fs"
	"os"
	fp "path/filepath"
	"runtime"
	"strings"
	"testing"
)

// ADir tests the behavior of a Dir-instance representing either a
// temporary directory obtained from a testing instance's FS-method or a
// package's "testdata"-directory.  It provides filesystem operations
// relative to that directory, i.e. with that directory as root.
//
//	func(s *MySuite) My_suite_test(t *T) {
//	    td := t.FS().Tmp()
//	}
type ADir struct{ Suite }

func (s *ADir) SetUp(t *T) { t.Parallel() }

func (s *ADir) Has_reported_path_created(t *T) {
	_, err := os.Stat(t.FS().Tmp().Path())
	t.True(err == nil)
}

func (s *ADir) Makes_nested_directories_of_a_series_of_names(t *T) {
	exp := []string{"a", "path", "of", "directories"}

	nested, _ := t.FS().Tmp().Mk(exp[0], exp[1:]...)

	t.True(strings.HasSuffix(nested.Path(), fp.Join(exp...)))
	_, err := os.Stat(nested.Path())
	t.True(err == nil)
}

func (s *ADir) Can_undo_a_directory_creation(t *T) {
	exp := []string{"a", "path", "of", "directories"}
	nested, undo := t.FS().Tmp().Mk(exp[0], exp[1:]...)

	path := nested.Path()
	undo()

	_, err := os.Stat(path)
	t.ErrIs(err, os.ErrNotExist)
}

func (s *ADir) Fatales_if_directory_creation_fails(t *T) {
	exp := []string{"a", "path", "of", "directories"}
	fx, failed := NewFS(t), false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().MkdirAll(func(_ string, _ fs.FileMode) error {
		return errors.New("mock-err")
	})

	fx.Tmp().Mk(exp[0], exp[1:]...)

	t.True(failed)
}

func (s *ADir) Panics_if_undoing_directory_creation_fails(t *T) {
	exp := []string{"a", "path", "of", "directories"}
	fx := NewFS(t)
	_, undo := fx.Tmp().Mk(exp[0], exp[1:]...)
	fx.Mock().RemoveAll(func(_ string) error {
		return errors.New("mock-err")
	})

	t.Panics(func() { undo() })
}

func (s *ADir) Creates_file_with_given_name_and_content(t *T) {
	td := t.FS().Tmp()
	expFl, expContent := "test.txt", []byte("the answer is 42\n")

	td.MkFile(expFl, expContent)

	_, err := os.Stat(fp.Join(td.Path(), expFl))
	t.FatalIfNot(t.True(err == nil))
	t.Eq(expContent, td.FileContent(expFl))
}

func (s *ADir) Can_undo_file_creation(t *T) {
	td := t.FS().Tmp()
	expFl, expContent := "test.txt", []byte("fearless\n")
	undo := td.MkFile(expFl, expContent)
	_, err := os.Stat(fp.Join(td.Path(), expFl))
	t.FatalIfNot(t.True(err == nil))

	undo()

	_, err = os.Stat(fp.Join(td.Path(), expFl))
	t.ErrIs(err, os.ErrNotExist)

}

func (s *ADir) Fatales_if_file_to_create_exists(t *T) {
	td := t.FS().Tmp()
	expFl, expContent, failed := "test.txt", []byte("fearless\n"), false

	td.MkFile(expFl, expContent)
	t.Mock().Canceler(func() { failed = true })
	td.MkFile(expFl, expContent)

	t.True(failed)
}

func (s *ADir) Fatales_if_file_write_fails(t *T) {
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

func (s *ADir) Panics_if_undoing_file_creation_fails(t *T) {
	fx := NewFS(t)
	expFl, expContent := "test.txt", []byte("fearless\n")
	undo := fx.Tmp().MkFile(expFl, expContent)
	fx.Mock().Remove(func(s string) error {
		return errors.New("mock-err")
	})

	t.Panics(func() { undo() })
}

func fxFile(t *T) (fx *FSfx, td *Dir, name string, content []byte) {
	name, content, fx = "test.txt", []byte("joyful\n"), NewFS(t)
	td = fx.Tmp()
	td.MkFile(name, content)
	return
}

func (s *ADir) Copies_file_to_other_dir(t *T) {
	_, td, expFl, expContent := fxFile(t)
	other, _ := td.Mk("other")

	td.CopyFl(expFl, other)

	t.Eq(expContent, td.FileContent(expFl))
	t.Eq(td.FileContent(expFl), other.FileContent(expFl))
}

func (s *ADir) Can_undo_file_copy(t *T) {
	_, td, expFl, _ := fxFile(t)
	other, _ := td.Mk("other")
	undo := td.CopyFl(expFl, other)

	undo()

	_, err := os.Stat(fp.Join(other.Path(), expFl))
	t.ErrIs(err, os.ErrNotExist)
}

func (s *ADir) Fatales_file_copy_if_file_read_fails(t *T) {
	fx, td, expFl, _ := fxFile(t)
	failed := false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().ReadFile(func(s string) ([]byte, error) {
		return nil, errors.New("mock-err")
	})

	td.CopyFl(expFl, fx.Tmp())

	t.True(failed)
}

func (s *ADir) Fatales_file_copy_if_file_write_fails(t *T) {
	fx, td, expFl, _ := fxFile(t)
	failed := false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().WriteFile(func(s string, b []byte, fm fs.FileMode) error {
		return errors.New("mock-err")
	})

	td.CopyFl(expFl, fx.Tmp())

	t.True(failed)
}

func (s *ADir) Panics_if_undoing_of_a_file_copy_fails(t *T) {
	fx, td, expFl, _ := fxFile(t)
	undo := td.CopyFl(expFl, fx.Tmp())
	fx.Mock().Remove(func(s string) error {
		return errors.New("mock-err")
	})

	t.Panics(undo)
}

func (s *ADir) Copies_its_file_structure_to_an_other_dir(t *T) {
	d1, _ := t.FS().Tmp().Mk("base")
	d2, _ := t.FS().Tmp().Mk("other")
	n1, _ := d1.Mk("nested")
	n1.MkFile("file.txt", []byte("108"))
	os.Symlink(d2.path, fp.Join(n1.path, "testLink"))

	d1.Copy(d2)

	t.True(d2.Child("base").Diff(d1))
}

func fxTwoDirs(t *T) (fx *FSfx, _, _, nested *Dir, file string) {
	fx = NewFS(t)
	d1, _ := fx.Tmp().Mk("base")
	d2, _ := fx.Tmp().Mk("other")
	n1, _ := d1.Mk("nested")
	n1.MkFile("file.txt", []byte("108"))
	return fx, d1, d2, n1, "file.txt"
}

func (s *ADir) Can_undo_a_directory_copy(t *T) {
	_, d1, d2, _, _ := fxTwoDirs(t)

	undo := d1.Copy(d2)
	cpy := d2.Child("base")
	_, err := os.Stat(cpy.Path())
	t.FatalOn(err)
	undo()

	_, err = os.Stat(cpy.Path())
	t.ErrIs(err, os.ErrNotExist)
}

var ErrMock = errors.New("mock-err")

func (s *ADir) Copy_fails_for_irregular_files_but_link(t *T) {
	// TODO: how to create an irregular file other than directory and
	// link

	// d1, _ := t.FS().Tmp().Mk("base")
	// d2, _ := t.FS().Tmp().Mk("other")
	// n1, _ := d1.Mk("nested")
	// n1.MkFile("file.txt", []byte("108"))
	// t.FatalOn(os.Chmod(fp.Join(n1.Path(), "file.txt"), 0644|os.ModeCharDevice))

	// undo := d1.Copy(d2)
	// cpy := d2.Child("base")
	// _, err := os.Stat(cpy.Path())
	// t.FatalOn(err)
	// undo()

	// _, err = os.Stat(cpy.Path())
	// t.ErrIs(err, os.ErrNotExist)
}

func (s *ADir) Fatales_copy_if_a_directory_cant_be_created(t *T) {
	fx, d1, d2, _, _ := fxTwoDirs(t)
	failed := false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().MkdirAll(func(s string, fm fs.FileMode) error {
		return ErrMock
	})

	d1.Copy(d2)
	t.True(failed)
}

func (s *ADir) Fatales_copy_if_a_file_cant_be_opened(t *T) {
	fx, d1, d2, _, _ := fxTwoDirs(t)
	failed := false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().Open(func(s string) (*os.File, error) {
		return nil, ErrMock
	})

	d1.Copy(d2)
	t.True(failed)
}

func (s *ADir) Fatales_copy_if_a_file_cant_be_created(t *T) {
	fx, d1, d2, _, _ := fxTwoDirs(t)
	failed := false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().Create(func(s string) (*os.File, error) {
		return nil, ErrMock
	})

	d1.Copy(d2)
	t.True(failed)
}

func (s *ADir) Fatales_copy_if_a_file_cant_be_copied(t *T) {
	fx, d1, d2, _, _ := fxTwoDirs(t)
	failed := false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().Copy(func(w io.Writer, r io.Reader) (int64, error) {
		return 0, errors.New("mock-err")
	})

	d1.Copy(d2)
	t.True(failed)
}

func (s *ADir) Fatales_copy_if_a_file_s_mod_cant_be_adapted(t *T) {
	fx, d1, d2, _, _ := fxTwoDirs(t)
	failed := false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().Chmod(func(s string, fm fs.FileMode) error {
		return errors.New("mock-err")
	})

	d1.Copy(d2)
	t.True(failed)
}

func (s *ADir) Panics_if_undoing_coping_fails(t *T) {
	fx, d1, d2, _, _ := fxTwoDirs(t)
	fx.Mock().RemoveAll(func(s string) error {
		return ErrMock
	})
	undo := d1.Copy(d2)

	t.Panics(undo)
}

func TestADir(t *testing.T) {
	t.Parallel()
	Run(&ADir{}, t)
}

type ADirDiff struct{ Suite }

func (s *ADirDiff) Fails_if_given_directories_have_different_bases(t *T) {
	d1, d2 := t.FS().Tmp(), t.FS().Tmp()
	t.FatalIfNot(t.False(fp.Base(d1.Path()) == fp.Base(d2.Path())))

	t.False(d1.Diff(d2))
}

func (s *ADirDiff) Fails_if_dirs_have_not_the_same_amounts_of_files(
	t *T,
) {
	d1, _ := t.FS().Tmp().Mk("base")
	d2, _ := t.FS().Tmp().Mk("base")
	d1.MkFile("not_in.txt", []byte("compassion"))

	t.False(d1.Diff(d2))
}

func (s *ADirDiff) Fails_if_directories_have_not_the_same_files(t *T) {
	d1, _ := t.FS().Tmp().Mk("base")
	d2, _ := t.FS().Tmp().Mk("base")
	d1.MkFile("not_in.txt1", []byte("compassion"))
	d2.MkFile("not_in.txt2", []byte("compassion"))

	t.False(d1.Diff(d2))
}

func (s *ADirDiff) Fails_having_common_files_with_different_mode(t *T) {
	d1, _ := t.FS().Tmp().Mk("base")
	d2, _ := t.FS().Tmp().Mk("base")
	d1.MkFile("file.txt", []byte("compassion"))
	d2.MkFile("file.txt", []byte("compassion"))
	t.FatalOn(os.Chmod(fp.Join(d1.Path(), "file.txt"), 0622))

	t.False(d1.Diff(d2))
}

func (s *ADirDiff) Fails_having_common_files_with_different_size(t *T) {
	d1, _ := t.FS().Tmp().Mk("base")
	d2, _ := t.FS().Tmp().Mk("base")
	d1.MkFile("file.txt", []byte("compassion"))
	d2.MkFile("file.txt", []byte("wisdom"))

	t.False(d1.Diff(d2))
}

func (s *ADirDiff) Fatales_if_nested_dir_cant_be_red(t *T) {
	fx, fatales := NewFS(t), false
	d1, _ := fx.Tmp().Mk("base")
	d2, _ := fx.Tmp().Mk("base")
	d1.Mk("nested")
	d2.Mk("nested")
	t.Mock().Canceler(func() { fatales = true })
	fx.Mock().ReadDir(func(s string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{}, errors.New("mock-err")
	})

	d1.Diff(d2)

	t.True(fatales)
}

type dirInfoMock struct{}

func (i *dirInfoMock) Name() string      { panic("not implemented") }
func (i *dirInfoMock) IsDir() bool       { panic("not implemented") }
func (i *dirInfoMock) Type() fs.FileMode { panic("not implemented") }
func (i *dirInfoMock) Info() (fs.FileInfo, error) {
	return nil, errors.New("mock-err")
}

func (s *ADirDiff) Fatales_if_nested_dir_s_file_info_cant_be_obtained(
	t *T,
) {
	fx, fatales := NewFS(t), false
	d1, _ := fx.Tmp().Mk("base")
	d2, _ := fx.Tmp().Mk("base")
	d1.Mk("nested")
	d2.Mk("nested")
	t.Mock().Canceler(func() { fatales = true })
	fx.Mock().ReadDir(func(s string) ([]fs.DirEntry, error) {
		return []fs.DirEntry{&dirInfoMock{}}, nil
	})

	d1.Diff(d2)

	t.True(fatales)
}

func (s *ADirDiff) Passes_if_non_of_the_above(t *T) {
	d1, _ := t.FS().Tmp().Mk("base")
	d2, _ := t.FS().Tmp().Mk("base")
	n1, _ := d1.Mk("nested")
	n2, _ := d2.Mk("nested")
	n1.MkFile("file.txt", []byte("108"))
	n2.MkFile("file.txt", []byte("108"))

	t.True(d1.Diff(d2))
}

func (s *ADirDiff) Fatales_if_the_dirs_file_info_cant_be_obtained(t *T) {
	fx := NewFS(t)
	d1, _ := fx.Tmp().Mk("base")
	d2, _ := fx.Tmp().Mk("base")
	fail := 0
	t.Mock().Canceler(func() {
		fail++
		if fail == 2 {
			t.Mock().Reset()
		}
	})
	fx.Mock().Stat(func(s string) (fs.FileInfo, error) {
		return nil, errors.New("mock-err")
	})

	t.Panics(func() { d1.Diff(d2) })

	t.Eq(2, fail)
}

func TestADirDiff(t *testing.T) {
	t.Parallel()
	Run(&ADirDiff{}, t)
}

var cllDir = func() string {
	_, f, _, ok := runtime.Caller(0)
	if !ok {
		panic("can't determine package directory")
	}
	return fp.Dir(f)
}()

// Testdata tests the behavior of FS.Data providing the
// testdata-directory for the calling package.  Since the tests of this
// suite potentially manipulate the same directory the tests can not run
// in parallel.
type Testdata struct{ Suite }

func (s *Testdata) fxMoveTestdata(t *T) func() {
	tmp := t.FS().Tmp()
	td, _ := t.FS().Data()
	td.Copy(tmp)
	if !tmp.Child(fp.Base(td.Path())).Diff(td) {
		panic("consistency error: copied directory should be " +
			"equal to source.")
	}
	os.RemoveAll(td.Path())
	return func() {
		dir, _ := t.FS().Dir(cllDir)
		os.RemoveAll(fp.Join(dir.Path(), "testdata"))
		tmp.Child("testdata").Copy(dir)
	}
}

func (s *Testdata) Directory_is_created_if_not_existing(t *T) {
	if _, err := os.Stat(fp.Join(cllDir, "testdata")); err == nil {
		defer s.fxMoveTestdata(t)()
	}
	_, err := os.Stat(fp.Join(cllDir, "testdata"))
	t.FatalIfNot(t.ErrIs(err, os.ErrNotExist))

	td, _ := t.FS().Data()
	_, err = os.Stat(fp.Join(cllDir, "testdata"))
	t.FatalIfNot(t.True(err == nil))
	t.Eq(fp.Join(cllDir, "testdata"), td.Path())
}

func (s *Testdata) Creation_can_be_undone(t *T) {
	if _, err := os.Stat(fp.Join(cllDir, "testdata")); err == nil {
		defer s.fxMoveTestdata(t)()
	}
	_, undo := t.FS().Data()
	_, err := os.Stat(fp.Join(cllDir, "testdata"))
	t.FatalIfNot(t.True(err == nil))

	undo()

	_, err = os.Stat(fp.Join(cllDir, "testdata"))
	t.FatalIfNot(t.ErrIs(err, os.ErrNotExist))
}

func (s *Testdata) Undo_is_omitted_if_test_data_exists_already(t *T) {
	if _, err := os.Stat(fp.Join(cllDir, "testdata")); err != nil {
		os.MkdirAll(fp.Join(cllDir, "testdata"), 0711)
		defer func() {
			if err := os.RemoveAll(fp.Join(cllDir, "testdata")); err != nil {
				panic(err)
			}
		}()
	}

	_, undo := t.FS().Data()

	t.True(undo == nil)
}

func (s *Testdata) Fatales_if_directory_creation_fails(t *T) {
	if _, err := os.Stat(fp.Join(cllDir, "testdata")); err == nil {
		defer s.fxMoveTestdata(t)()
	}
	fx, failed := NewFS(t), false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().MkdirAll(func(_ string, _ fs.FileMode) error {
		return ErrMock
	})
	defer fx.Mock().Reset()

	fx.Data()

	t.True(failed)
}

func (s *Testdata) Fatales_if_caller_cant_be_determined(t *T) {
	if _, err := os.Stat(fp.Join(cllDir, "testdata")); err == nil {
		defer s.fxMoveTestdata(t)()
	}
	fx, failed := NewFS(t), false
	t.Mock().Canceler(func() { failed = true })
	fx.Mock().Caller(func(i int) (uintptr, string, int, bool) {
		return 0, "", 0, false
	})
	defer fx.Mock().Reset()

	fx.Data()

	t.True(failed)
}

func (s *Testdata) Panics_if_testdata_creation_undoing_fails(t *T) {
	if _, err := os.Stat(fp.Join(cllDir, "testdata")); err == nil {
		defer s.fxMoveTestdata(t)()
	}
	fx := NewFS(t)
	fx.Mock().RemoveAll(func(s string) error {
		return ErrMock
	})
	defer fx.Mock().Reset()

	_, undo := fx.Data()

	t.Panics(undo)
}

func TestTestdata(t *testing.T) {
	Run(&Testdata{}, t)
}
