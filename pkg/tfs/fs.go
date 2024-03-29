// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/*
Package tfs provides complex filesystem operations for testing.  I.e.
an operation is usually at least comprised of two (potentially
failing) standard lib filesystem operations.  Error handling is
fully replaced by the assumption that it's not worth to run a
test whose filesystem operations fail since the typical use case is a
fixture setup.  Hence if a filesystem operation fails it fatales the
associated test.

An FS instance is usually obtained from a gounit.T instance:

	import (
		"testing"

		"github.com/slukits/gounit"
	)

	type MySuite { gounit.Suite }

	func (s *MySuite) Suite_test(t *gounit.T) {
		td := t.FS().Tmp() // create temp-dir
		gd := t.FS().Data().Child("golden_dir") // golden_dir from testdata
		gd.Copy(td) // copied to temp dir
		t.FatalIfNot(t.True(td.Child("golden_dir").Eq(gd)))
		// ...
	}

	func TestMySuite(t *testing.T) { Run(&MySuite{}, t) }
*/
package tfs

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	fp "path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
	"time"
)

// Tester summarizes and abstracts what a [FS] instance needs from an
// associated test's testing instance.
type Tester interface {
	Fatal(...interface{})
	Fatalf(string, ...interface{})
	GoT() *testing.T
	FS() *FS
}

// FS provides directory relative filesystem operation specifically for
// testing, e.g.  without error handling, preset file/dir-mod,
// constructors for unique temp-dir as well as testdata directory, and
// with undo functionality.  In general failing file system operations
// fatal associated testing instance while failing undo function-calls
// panic.
type FS struct {
	t     Tester
	td    *Dir
	tools *fsTools
}

func New(t Tester) *FS {
	return &FS{t: t, tools: defaultFSTools}
}

// tls provides the file system tools to created Dir and TmpDir instances.
func (fs *FS) tls() *fsTools { return fs.tools }

// Data returns the callers testdata directory.  Associated testing
// instance fatales if the directory doesn't exist and can't be created.
// Returned undo function is nil in case the testdata directory already
// existed, i.e. also for subsequent calls to Data after it has been
// created at the first call.  Returned undo function panics if its
// execution fails.
func (fs *FS) Data() (_ *Dir, undo func()) {
	if fs.td != nil {
		if _, err := os.Stat(fs.td.path); err == nil {
			return fs.td, nil
		}
	}
	_, f, _, ok := fs.tools.Caller(1)
	if !ok {
		fs.t.Fatal("gounit: fs: testdata: can't determine caller")
	}

	return fs.Dir(fp.Join(fp.Dir(f), "testdata"))
}

// Dir wraps given path in a directory which is created if not existing.
// In the later case an undo function is returned to undo the directory
// creation.  Dir fatales associated test on any fs-error.   Returned
// undo panics if its execution fails.
func (fs *FS) Dir(path string) (_ *Dir, undo func()) {
	created := false
	if _, err := fs.tools.Stat(path); err != nil {
		if err := fs.tools.MkdirAll(path, 0711); err != nil {
			fs.t.Fatal("gounit: fs: testdata: create: %v", err)
		}
		created = true
	}
	fs.td = &Dir{t: fs.t, path: path, fs: fs.tls}

	if !created {
		return fs.td, nil
	}

	return fs.td, func() {
		if err := fs.tools.RemoveAll(path); err != nil {
			panic(err)
		}
	}
}

// Tmp creates a new unique temporary directory bound to associated
// testing instance.  Associated testing instance fatales if the temp
// directory creation fails.
func (fs *FS) Tmp() *Dir {
	return &Dir{t: fs.t, path: fs.t.GoT().TempDir(), fs: fs.tls}
}

// Dir provides file system operations inside its path, which is
// typically nested in either a temporary directory or the package's
// testdata directory.  It replaces error handling by failing the test.
// The zero value of a Dir instance is *NOT* usable.  Use [FS.Data],
// [FS.Dir] or [FS.Tmp] to obtain a Dir-instance:
//
//	func (s *MySuite) Suite_test(t *gounit.T) {
//	    td := t.FS().Tmp() // create temporary Dir-instance
//	    // ...
//	}
type Dir struct {
	t    Tester
	fs   func() *fsTools
	path string
}

// Path returns the directory's path.
func (d *Dir) Path() string { return d.path }

// Child returns a Dir in given directory d with given name.  Child
// fatales associated test if child's file info can't be obtained or if
// child is not a directory.
func (d *Dir) Child(name string) *Dir {
	stt, err := d.fs().Stat(fp.Join(d.path, name))
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: child: %s: %v", name, err)
	}
	if !stt.IsDir() {
		d.t.Fatalf("gounit: fs: dir: child: %s: is no directory", name)
	}
	return &Dir{t: d.t, fs: d.fs, path: fp.Join(d.path, name)}
}

type Pather interface{ Path() string }

// Copy copies given directory d to given path toDir and returns a
// function to undo this operation.  Note next to directories only
// regular files and symlinks are supported.  Copy fatales associated
// test on other irregular files as well as on any failing filesystem
// operation.  Returned undo panics if its execution fails.
func (d *Dir) Copy(toDir Pather) (undo func()) {

	err := d.fs().Walk(d.path, func(
		path string, info fs.FileInfo, err error,
	) error {

		if err != nil {
			return err
		}

		dstPath := fp.Join(
			toDir.Path(), strings.TrimPrefix(path, fp.Dir(d.path)))

		if info.IsDir() {
			err := d.fs().MkdirAll(dstPath, info.Mode())
			if err != nil {
				return err
			}
			return nil // means recursive
		}

		if !info.Mode().IsRegular() {
			switch info.Mode().Type() & os.ModeType {
			case os.ModeSymlink:
				link, err := os.Readlink(path)
				if err != nil {
					return err
				}
				return d.fs().Symlink(link, dstPath)
			}
			return fmt.Errorf(
				"gounit: fs: dir: copy: can't handle file: %s",
				path,
			)
		}

		src, err := d.fs().Open(path)
		if err != nil {
			return err
		}
		defer src.Close()

		dst, err := d.fs().Create(dstPath)
		if err != nil {
			return err
		}
		defer dst.Close()

		if _, err = d.fs().Copy(dst, src); err != nil {
			return err
		}
		if err := d.fs().Chmod(dstPath, info.Mode()); err != nil {
			return err
		}

		return err
	})

	if err != nil {
		d.t.Fatalf("gounit: fs: dir: copy: %v", err)
	}

	return func() {
		dir := fp.Join(toDir.Path(), fp.Base(d.path))
		if err := d.fs().RemoveAll(dir); err != nil {
			panic(err)
		}
	}
}

// Eq return true if given dir d's and given path p's last directories
// have the same name, both contain the same directory structure with
// the same files whereas two files are considered the same iff they
// have the same name, the same size and the same mode.  I.e. the
// content of two files is not compared.  Otherwise Eq returns false.
// Eq fatales associated test if any of the executed file system
// operations fails.  Note this implementation could be more efficient
// by avoiding the FileInfo calculation in case of directories.
func (d *Dir) Eq(p Pather) bool {

	src, dest := fp.Dir(d.path), fp.Dir(p.Path())
	srcFI, err := d.fs().Stat(d.path)
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: diff: %v", err)
	}
	dstFI, err := d.fs().Stat(p.Path())
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: diff: %v", err)
	}
	// "stack" for sources relative to src
	ss := map[string]fs.FileInfo{srcFI.Name(): srcFI}
	// "stack" for destinations relative to dest
	dd := map[string]fs.FileInfo{dstFI.Name(): dstFI}

	for len(ss) > 0 {
		if len(ss) != len(dd) {
			return false
		}
		for k := range ss { // check files/dirs for equality
			if _, ok := dd[k]; !ok {
				return false
			}
			if ss[k].Mode() != dd[k].Mode() {
				return false
			}
			if ss[k].Size() != dd[k].Size() {
				return false
			}
		}
		dirs := []string{}
		for k := range ss { // remove all files
			if ss[k].IsDir() {
				dirs = append(dirs, k)
			}
			delete(ss, k)
			delete(dd, k)
		}
		for _, dr := range dirs { // replace directories with content
			d.replaceDirWithContent(src, dr, &ss)
			d.replaceDirWithContent(dest, dr, &dd)
		}
	}

	return true
}

func (d *Dir) replaceDirWithContent(
	parent, dir string, m *map[string]fs.FileInfo,
) {
	delete(*m, dir)
	ee, err := d.fs().ReadDir(fp.Join(parent, dir))
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: diff: %v", err)
	}
	for _, e := range ee {
		fi, err := e.Info()
		if err != nil {
			d.t.Fatalf("gounit: fs: dir: diff: %v", err)
			return
		}
		(*m)[fp.Join(dir, e.Name())] = fi
	}
}

// Mk crates a new directory inside given directory d's path by
// combining given strings dir and path to a (relative) path.  The
// returned function removes the root directory of given path and resets
// returned Dir instance.  It fails associated test in case the
// directory creation fails.  Returned undo panics if its execution fails.
func (d *Dir) Mk(dir string, path ...string) (_ *Dir, undo func()) {
	_path := fp.Join(append([]string{d.path, dir}, path...)...)
	if err := d.fs().MkdirAll(_path, 0711); err != nil {
		d.t.Fatalf("gounit: fs: dir: create: %v", err)
	}
	new := &Dir{t: d.t, path: _path, fs: d.fs}
	return new, func() {
		new.t = nil
		new.path = ""
		if err := d.fs().RemoveAll(fp.Join(d.path, dir)); err != nil {
			panic(fmt.Sprintf("gounit: fs: dir: reset: %v", err))
		}
	}
}

// Rm removes in given directory d given relatives directory rel with
// all its files and sub-directories.  Rm fatales associated test the
// removal fails.
func (d *Dir) Rm(rel string) {
	d.t.GoT().Helper()
	if err := d.fs().RemoveAll(fp.Join(d.path, rel)); err != nil {
		d.t.Fatalf("gounit: fs: dir: remove: '%s': %v", rel, err)
	}
}

// MkFile adds to given directory d a new file (mod 0644) with given
// name and content and returns an undo function.  MkFile fatales
// associated test if the file already exists or os.WriteFile fails.
// Returned undo panics if its execution fails.
func (d *Dir) MkFile(name string, content []byte) (undo func()) {
	d.t.GoT().Helper()

	fl := fp.Join(d.path, name)
	if _, err := d.fs().Stat(fl); err == nil {
		d.t.Fatalf("gounit: fs: dir: add file: already exists")
	}

	if err := d.fs().WriteFile(fl, []byte(content), 0644); err != nil {
		d.t.Fatalf("gounit: fs: dir: add file: write: %v", err)
	}

	return func() {
		if err := d.fs().Remove(fl); err != nil {
			panic(fmt.Sprintf("fx: add file: reset: %v", err))
		}
	}
}

// FileCopy copies the content of given file from given directory d to
// given path toDir.  FileCopy fatales associated test if ReadFile or
// WriteFile fails.  Returned undo function removes the copy and panics
// if its execution fails.
func (d *Dir) FileCopy(file string, toDir Pather) (undo func()) {
	d.t.GoT().Helper()

	bb, err := d.fs().ReadFile(fp.Join(d.path, file))
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: copy: read: %s: %v", file, err)
	}
	err = d.fs().WriteFile(fp.Join(toDir.Path(), file), bb, 0644)
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: copy: write: %s: %v", file, err)
	}
	return func() {
		if err := d.fs().Remove(fp.Join(toDir.Path(), file)); err != nil {
			panic(err)
		}
	}
}

// FileContent joins given directory d with given file name relName and returns
// its content.  FileContent fatales if it cant be read.
func (d *Dir) FileContent(relName string) []byte {
	d.t.GoT().Helper()

	bb, err := d.fs().ReadFile(fp.Join(d.path, relName))
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: file-copy: read: %s: %v", relName, err)
	}
	return bb
}

// WriteContent replaces the content of the file with given name in
// given directory with given bytes.  WriteContent fatales if writing
// the file fails.
func (d *Dir) WriteContent(relName string, bb []byte) {
	d.t.GoT().Helper()
	err := d.fs().WriteFile(fp.Join(d.Path(), relName), bb, 0644)
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: write content: %s: %v", relName, err)
	}
}

// FileMod returns a directory-file relName's modification time.  It fatales
// associated test iff stats can't be obtained.
func (d *Dir) FileMod(relName string) time.Time {
	stat, err := d.fs().Stat(fp.Join(d.path, relName))
	if err != nil {
		d.t.Fatalf("gounit: fs: dir: file: mod-stat: %s: %v", relName, err)
	}
	return stat.ModTime()
}

// Touch sets given directory-file's modification (access) time to the
// current time; let d be a Dir-instance:
//
//	mt := d.FileMod("my_file")
//	time.Sleep(1*time.Millisecond)
//	d.Touch("my_file")
//	fmt.Println(mt.Before(d.FileMod("my_file")) // => "true"
//
// Touch fatales associated test iff given file doesn't exist in given
// directory or if its modification time can't be updated.
func (d *Dir) Touch(relName string) {
	fileName := fp.Join(d.path, relName)
	if _, err := os.Stat(fileName); err != nil {
		d.t.Fatalf("gounit: fs: dir: touch: %s: %v", relName, err)
	}
	now := time.Now()
	if err := os.Chtimes(fileName, now, now); err != nil {
		d.t.Fatalf("gounit: fs: dir: touch: %s: %v", relName, err)
	}
}

// MkMod adds to given directory d a go.mod file with given module name.
// It fatales if subsequent [Dir.MkFile] call fatales.  Returned undo
// panics if its execution fails.
func (d *Dir) MkMod(module string) (undo func()) {
	d.t.GoT().Helper()
	return d.MkFile("go.mod", []byte(fmt.Sprintf("module %s", module)))
}

// MkTidy tries to make sure that a temporary go module has all packages
// references needed by its packages.  Once successfully executed go.mod
// and go.sum are cached in the users caching path's gounit directory.
// This cached files are considered stale every 24 hours.
func (d *Dir) MkTidy() {
	d.t.GoT().Helper()
	mdlName := d.moduleName()
	if mdlName == "" {
		d.t.Fatal("gounit: fs: tmp-dir: mk-tidy: missing module name")
	}
	d.mkGoModSum(mdlName)
}

// HasCachedModSum returns true if current directory is a go module and
// there are go.mod and go.sum files cached for this module; otherwise
// false.
func (d *Dir) HasCachedModSum() bool {
	mdlName := d.moduleName()
	if mdlName == "" {
		return false
	}
	_, _, ok := goModSumFromCache(d.fs(), mdlName)

	return ok
}

// mkGoModSum creates needed go.mod and go.sum file for give module in
// its root directory.  mkGoModSum first tries to obtain these files
// from the cache; is that no possible it ties to created them by
// executing "go mod tidy"; updated go.mod and created go.sum file are
// tried to be cached for the next call.  Note we can not depend on
// functioning caching since it will never work for example on github
// during a workflow which runs all tests.
func (d *Dir) mkGoModSum(module string) {
	modFl := fp.Join(d.path, "go.mod")
	sumFl := fp.Join(d.path, "go.sum")
	mod, sum, ok := goModSumFromCache(d.fs(), module)
	if ok {
		if err := d.fs().WriteFile(modFl, mod, 0644); err != nil {
			d.t.Fatal("gounit: fs: tmp-dir: tidy: write go.mod: %v",
				err)
		}
		if err := d.fs().WriteFile(sumFl, sum, 0644); err != nil {
			d.t.Fatal("gounit: fs: tmp-dir: tidy: write go.sum: %v",
				err)
		}
		return
	}
	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = d.path
	if stdout, err := cmd.CombinedOutput(); err != nil {
		d.t.Fatalf("gounit: fs: tmp-dir: tidy: go mod tidy: %v:\n%v",
			err, stdout)
	}
	bbMod, err := d.fs().ReadFile(modFl)
	if err != nil {
		return
	}
	bbSum, err := d.fs().ReadFile(sumFl)
	if err != nil {
		return
	}
	goModSumToCache(d.fs(), module, bbMod, bbSum)
}

// goModSumFromCache tries to read given modules go.mod and go.sum from
// the user's caching path's gounit directory and returning.  In case of
// success return ok is true otherwise it is false.
func goModSumFromCache(
	fs *fsTools, module string,
) (bbMod, bbSum []byte, ok bool) {
	cch, err := os.UserCacheDir()
	if err != nil {
		return nil, nil, false
	}
	modCache := fp.Join(cch, "gounit", module, "go.mod")
	sumCache := fp.Join(cch, "gounit", module, "go.sum")
	stt, err := fs.Stat(modCache)
	if err != nil || time.Since(stt.ModTime()) > 24*time.Hour {
		return nil, nil, false
	}
	stt, err = fs.Stat(sumCache)
	if err != nil || time.Since(stt.ModTime()) > 24*time.Hour {
		return nil, nil, false
	}
	bbMod, err = fs.ReadFile(modCache)
	if err != nil {
		return nil, nil, false
	}
	bbSum, err = fs.ReadFile(sumCache)
	if err != nil {
		return nil, nil, false
	}
	return bbMod, bbSum, true
}

func goModSumToCache(fs *fsTools, module string, bbMod, bbSum []byte) {
	cch, err := os.UserCacheDir()
	if err != nil {
		return
	}
	cacheDir := fp.Join(cch, "gounit", module)
	if _, err := fs.Stat(cacheDir); err != nil {
		if err := fs.MkdirAll(cacheDir, 0711); err != nil {
			return
		}
	}
	modCache := fp.Join(cacheDir, "go.mod")
	sumCache := fp.Join(cacheDir, "go.sum")
	if err := fs.WriteFile(modCache, bbMod, 0644); err != nil {
		return
	}
	fs.WriteFile(sumCache, bbSum, 0644)
}

func (d *Dir) moduleName() string {
	goMod, err := d.fs().Open(fp.Join(d.path, "go.mod"))
	if err != nil {
		d.t.Fatalf("gounit: fs: tmp-dir: module-name: %v", err)
	}
	defer goMod.Close()

	scanner := bufio.NewScanner(goMod)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "module ") {
			continue
		}
		if err := scanner.Err(); err != nil {
			d.t.Fatalf("gounit: fs: tmp-dir: module-name: %v", err)
		}
		return line[len("module "):]
	}
	return ""
}

var rePkgComment = regexp.MustCompile(`(?s)^(\s*?\n|// .*?\n|/\*.*\*/)*`)

// MkPkgFile adds a file in given directory d prefixing given content
// with a package declaration and suffixing given file name with ".go"
// if missing.  MkPkgFile fatales associated test on failing
// fs-operations.  Returned undo panics if its execution fails.
func (d *Dir) MkPkgFile(name string, content []byte) (undo func()) {
	pkg := fp.Base(d.path)
	if !bytes.Contains(content, []byte(fmt.Sprintf("package %s", pkg))) {
		content = rePkgComment.ReplaceAll(
			content, []byte(fmt.Sprintf("$1\npackage %s\n\n", pkg)))
		content = bytes.TrimLeft(content, "\n")
	}
	if !strings.HasSuffix(name, ".go") {
		name = fmt.Sprintf("%s.go", name)
	}
	return d.MkFile(name, []byte(content))
}

// MkPkgTest adds a test file in given directory d prefixing given
// content with a package declaration and suffixing given file name with
// "_test.go" if missing.  MkPkgTest fatales associated test on failing
// fs-operations.  Returned undo panics if its execution fails.
func (d *Dir) MkPkgTest(name string, content []byte) (undo func()) {
	if !strings.HasSuffix(name, "_test.go") {
		name = fmt.Sprintf("%s%s", name, "_test.go")
	}
	return d.MkPkgFile(name, content)
}

// CWD changes the current working directory to given directory and
// returns a function to undo this change.  CWD fatales associated
// testing instance if the working directory change fails.  Returned
// undo panics if its execution fails.
func (d *Dir) CWD() (undo func()) {
	wd, err := d.fs().Getwd()
	if err != nil {
		d.t.Fatalf("gounit: fs: tmp-dir: cwd: get: %v", err)
		return nil
	}

	if err := d.fs().Chdir(d.path); err != nil {
		d.t.Fatalf("gounit: fs: tmp-dir: cwd: %v", err)
		return nil
	}

	return func() {
		if err := d.fs().Chdir(wd); err != nil {
			panic(fmt.Sprintf("gounit: fs: tmp-dir: cwd: reset: %v", err))
		}
	}
}

// fsTools are the functions for potentially failing file system
// operation which are used by Dir and TmpDir instances.
type fsTools struct {

	// Stat defaults to and has the semantics of os.Stat
	Stat func(string) (fs.FileInfo, error)

	// Getwd defaults to and has the semantics of os.Getwd
	Getwd func() (string, error)

	// Chdir defaults to and has the semantics of os.Chdir
	Chdir func(string) error

	// Mkdir defaults to and has the semantics of os.Mkdir
	Mkdir func(string, fs.FileMode) error

	// MkdirAll defaults to and has the semantics of os.MkdirAll
	MkdirAll func(string, fs.FileMode) error

	// Remove defaults to and has the semantics of os.Remove
	Remove func(string) error

	// RemoveAll defaults to and has the semantics of os.RemoveAll
	RemoveAll func(string) error

	// Symlink defaults to and has the semantics of os.Symlink
	Symlink func(string, string) error

	// Open defaults to and has the semantics of os.Open
	Open func(string) (*os.File, error)

	// Create defaults to and has the semantics of os.Create
	Create func(string) (*os.File, error)

	// ReadDir defaults to and has the semantics of os.ReadDir
	ReadDir func(string) ([]fs.DirEntry, error)

	// ReadFile defaults to and has the semantics of os.ReadFile
	ReadFile func(string) ([]byte, error)

	// WriteFile defaults to and has the semantics of os.WriteFile
	WriteFile func(string, []byte, fs.FileMode) error

	// Chmod defaults to and has the semantics of os.Chmod
	Chmod func(string, fs.FileMode) error

	// Copy defaults to and has the semantics of io.Copy
	Copy func(io.Writer, io.Reader) (int64, error)

	// Walk defaults to and has the semantics of filepath.Walk
	Walk func(string, fp.WalkFunc) error

	// Caller default to and has the semantics of runtime.Caller
	Caller func(int) (uintptr, string, int, bool)
}

func (t *fsTools) copy() *fsTools {
	return &fsTools{
		Stat:      t.Stat,
		Getwd:     t.Getwd,
		Chdir:     t.Chdir,
		Mkdir:     t.Mkdir,
		MkdirAll:  t.MkdirAll,
		Remove:    t.Remove,
		RemoveAll: t.RemoveAll,
		Symlink:   t.Symlink,
		Open:      t.Open,
		Create:    t.Create,
		ReadDir:   t.ReadDir,
		ReadFile:  t.ReadFile,
		WriteFile: t.WriteFile,
		Chmod:     t.Chmod,
		Copy:      t.Copy,
		Walk:      t.Walk,
		Caller:    t.Caller,
	}
}

var defaultFSTools = func() *fsTools {
	return &fsTools{
		Stat:      os.Stat,
		Getwd:     os.Getwd,
		Chdir:     os.Chdir,
		Mkdir:     os.Mkdir,
		MkdirAll:  os.MkdirAll,
		Remove:    os.Remove,
		RemoveAll: os.RemoveAll,
		Symlink:   os.Symlink,
		Open:      os.Open,
		Create:    os.Create,
		ReadDir:   os.ReadDir,
		ReadFile:  os.ReadFile,
		WriteFile: os.WriteFile,
		Chmod:     os.Chmod,
		Copy:      io.Copy,
		Walk:      fp.Walk,
		Caller:    runtime.Caller,
	}
}()
