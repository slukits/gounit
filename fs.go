// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// FS provides filesystem operation specifically for testing, e.g.
// without error handling, preset file/dir-mod, restricted to temporary
// or testdata directories with undo functionality.  In general failing
// file system operations fatal associated testing instance; failing
// undo function calls panic.
type FS struct {
	t  *T
	td *Dir
}

// Data returns the callers testdata directory (Note not t.FS() but the
// caller one before ;).  Associated testing instance fatales if the
// directory doesn't exist and can't be created.
func (fs *FS) Data() *Dir {
	if fs.td == nil {
		_, f, _, ok := runtime.Caller(2)
		if !ok {
			fs.t.Fatal("gounit: fs: testdata: can't determine caller")
		}
		tdDir := filepath.Join(filepath.Dir(f), "testdata")
		if _, err := os.Stat(tdDir); err != nil {
			if err := os.Mkdir(tdDir, 0711); err != nil {
				fs.t.Fatal("gounit: fs: testdata: create: %v", err)
			}
		}
		fs.td = &Dir{t: fs.t, path: tdDir}
	}
	return fs.td
}

// Tmp creates a new unique temporary directory bound to associated
// testing instance.  Associated testing instance fatales if the temp
// directory creation fails.
func (fs *FS) Tmp() *TmpDir {
	return &TmpDir{Dir: Dir{t: fs.t, path: fs.t.GoT().TempDir()}}
}

// Dir provides file system operations inside its path, i.e. either a
// temporary directory or a a package's testdata directory.  It replaces
// error handling by failing the test using a Dir instance.  The zero
// value of a Dir instance is *NOT* usable.  Use gounit.T testing
// instance's [t.FS]-method to obtain a Dir-instance.
type Dir struct {
	t    *T
	path string
}

// Path returns the directory's directory, och, path.
func (d *Dir) Path() string { return d.path }

type Pather interface{ Path() string }

// Copy copies the content of given file from given directory to given
// Path().  Copy fatales associated testing instance if ReadFile or
// WriteFile fails.  Returned undo function removes the copy.
func (d *Dir) Copy(file string, toDir Pather) (undo func()) {
	bb, err := os.ReadFile(filepath.Join(d.path, file))
	if err != nil {
		d.t.Fatal("gounit: fs: dir: copy: read: %s: %v", file, err)
	}
	err = os.WriteFile(filepath.Join(d.path, file), bb, 0644)
	if err != nil {
		d.t.Fatal("gounit: fs: dir: copy: write: %s: %v", file, err)
	}
	return func() {
		if err := os.Remove(filepath.Join(d.path, file)); err != nil {
			panic(err)
		}
	}
}

// Mk crates a new directory inside given directory's path by combining
// given strings to a (relative) path.  The returned function removes
// the root directory of given path and resets returned Dir instance.
// It fails associated test in case the directory creation fails.  It
// panics in case the undo function fails.
func (d *Dir) Mk(path ...string) (_ *Dir, undo func()) {
	new := filepath.Join(append([]string{d.path}, path...)...)
	if err := os.MkdirAll(new, 0711); err != nil {
		d.t.Fatalf("gounit: fs: dir: create: %v", err)
	}
	dir := &Dir{t: d.t, path: new}
	return dir, func() {
		if len(path) == 0 {
			return
		}
		dir.t = nil
		dir.path = ""
		if err := os.RemoveAll(path[0]); err != nil {
			panic(fmt.Sprintf("gounit: fs: dir: reset: %v", err))
		}
	}
}

// MkFile adds to given directory a new file (mod 0644) with given name
// and given content.  MkFile fatales if the file already exists or
// os.WriteFile fails.  MkFile panics if reset fails.
func (d *Dir) MkFile(name, content string) (undo func()) {

	fl := filepath.Join(d.path, name)
	if _, err := os.Stat(fl); err == nil {
		d.t.Fatalf("gounit: fs: dir: add file: already exists")
	}

	if err := os.WriteFile(fl, []byte(content), 0644); err != nil {
		d.t.Fatalf("gounit: fs: dir: add file: write: %v", err)
	}

	return func() {
		if err := os.Remove(fl); err != nil {
			panic(fmt.Sprintf("fx: add file: reset: %v", err))
		}
	}
}

// TmpDir is a temporary directory created for testing.  It adds to
// features of its embedded Dir instance the possibility to make it the
// working directory.  Or to create a temporary go module with go mod
// file, packages and go mod tidy to resolve imports. (to save time
// go.mod and go.sum are cached if possible.)  The zero value of a Dir
// instance is *NOT* usable.  Use gounit.T testing instance's
// [t.FS]-method to obtain a TmpDir-instance.
type TmpDir struct {
	Dir
}

// MkMod adds to given directory a go.mod file with given module name.
// It fatales/panics iff subsequent [Dir.AddFile] call fatales/panics.
func (d *TmpDir) MkMod(module string) (reset func()) {
	os.UserCacheDir()
	return d.MkFile("go.mod", fmt.Sprintf("module %s", module))
}

// MkTidy tries to make sure that a temporary go module has all packages
// references needed by its packages.  Once successfully executed go.mod
// and go.sum are cached in the users caching path's gounit directory.
// This cached files are considered stale every 24 hours.
func (d *TmpDir) MkTidy() {
	mdlName := d.moduleName()
	if mdlName == "" {
		d.t.Fatal("gounit: fs: tmp-dir: mk-tidy: missing module name")
	}
	d.mkGoModSum(mdlName)
}

// goModSumFromCache tries to read given modules go.mod and go.sum from
// the user's caching path's gounit directory.  If that is not possible
// it is tried to created needed files by calling go mod tidy and to
// cache them.
func (d *TmpDir) mkGoModSum(module string) {
	modFl := filepath.Join(d.path, "go.mod")
	sumFl := filepath.Join(d.path, "go.sum")
	mod, sum, ok := goModSumFromCache(module)
	if ok {
		if err := os.WriteFile(modFl, mod, 0644); err != nil {
			d.t.Fatal("gounit: fs: tmp-dir: tidy: write go.mod: %v",
				err)
		}
		if err := os.WriteFile(sumFl, sum, 0644); err != nil {
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
	bbMod, err := os.ReadFile(modFl)
	if err != nil {
		return
	}
	bbSum, err := os.ReadFile(sumFl)
	if err != nil {
		return
	}
	goModSumToCache(module, bbMod, bbSum)
}

func goModSumFromCache(module string) (bbMod, bbSum []byte, ok bool) {
	cch, err := os.UserCacheDir()
	if err != nil {
		return nil, nil, false
	}
	modCache := filepath.Join(cch, "gounit", module, "go.mod")
	sumCache := filepath.Join(cch, "gounit", module, "go.sum")
	stt, err := os.Stat(modCache)
	if err != nil || time.Since(stt.ModTime()) > 24*time.Hour {
		return nil, nil, false
	}
	stt, err = os.Stat(sumCache)
	if err != nil || time.Since(stt.ModTime()) > 24*time.Hour {
		return nil, nil, false
	}
	bbMod, err = os.ReadFile(modCache)
	if err != nil {
		return nil, nil, false
	}
	bbSum, err = os.ReadFile(sumCache)
	if err != nil {
		return nil, nil, false
	}
	return bbMod, bbSum, true
}

func goModSumToCache(module string, bbMod, bbSum []byte) {
	cch, err := os.UserCacheDir()
	if err != nil {
		return
	}
	cacheDir := filepath.Join(cch, "gounit", module)
	if _, err := os.Stat(cacheDir); err != nil {
		if err := os.MkdirAll(cacheDir, 0711); err != nil {
			return
		}
	}
	modCache := filepath.Join(cacheDir, "go.mod")
	sumCache := filepath.Join(cacheDir, "go.sum")
	if err := os.WriteFile(modCache, bbMod, 0644); err != nil {
		return
	}
	os.WriteFile(sumCache, bbSum, 0644)
}

func (d *TmpDir) moduleName() string {
	goMod, err := os.Open(filepath.Join(d.path, "go.mod"))
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

// MkPkgFile adds a file with given content prefixing its content with a
// package declaration and suffixing given file name with ".go" if missing.
func (d *TmpDir) MkPkgFile(name, content string) (undo func()) {
	pkg := filepath.Base(d.path)
	if !strings.Contains(content, fmt.Sprintf("package %s", pkg)) {
		content = rePkgComment.ReplaceAllString(
			content, fmt.Sprintf("$1\npackage %s\n\n", pkg))
		content = strings.TrimLeft(content, "\n")
	}
	if !strings.HasSuffix(name, ".go") {
		name = fmt.Sprintf("%s.go", name)
	}
	return d.MkFile(name, content)
}

// MkPkgTest adds a test file with given content prefixing its content
// with a package declaration and suffixes "_test.go" to the name if
// missing.
func (d *TmpDir) MkPkgTest(name, content string) (undo func()) {
	if !strings.HasSuffix(name, "_test.go") {
		name = fmt.Sprintf("%s%s", name, "_test.go")
	}
	return d.MkPkgFile(name, content)
}

func (td *TmpDir) MkTmp(path ...string) (_ *TmpDir, undo func()) {
	new, undo := td.Mk(path...)
	return &TmpDir{Dir: *new}, undo
}

// CWD changes the current working directory to given temporary
// directory and returns a function to undo this change.  CWD fatales
// given testing instance if given directory if the working directory
// change fails.  It panics if the undo fails.
func (td *TmpDir) CWD() (undo func()) {
	wd, err := os.Getwd()
	if err != nil {
		td.t.Fatalf("gounit: fs: tmp-dir: cwd: get: %v", err)
	}

	if err := os.Chdir(td.path); err != nil {
		td.t.Fatalf("gounit: fs: tmp-dir: cwd: %v", err)
	}

	return func() {
		if err := os.Chdir(wd); err != nil {
			panic(fmt.Sprintf("gounit: fs: tmp-dir: cwd: reset: %v", err))
		}
	}
}
