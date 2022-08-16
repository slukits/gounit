// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// fs provides helpers to manipulate the file system to create
// test-fixtures.

package fx

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// NoTempDirErr is an error string indicating that a provided directory
// is not an expected temporary directory.
var NoTempDirErr = "is not temp-dir"

// MkPath creates in root given variadic series of directories in form of
// a descending path: filepath.Join(root, dd[0], ..., dd[len(dd)-1])
// MkPath fatales given testing instance if given root is not a temporary
// directory or the directories can not be created.  MkPath panics if
// reset fails.
func MkPath(t *testing.T, root string, dd ...string) (path string, reset func()) {
	if !strings.HasPrefix(root, os.TempDir()) {
		t.Fatalf("fx: dirs: root: %s", NoTempDirErr)
	}

	path = root
	for _, d := range dd {
		path = filepath.Join(path, d)
	}
	if err := os.MkdirAll(path, 0711); err != nil {
		t.Fatalf("fx: dirs: create: %v", err)
	}

	return path, func() {
		if len(dd) == 0 {
			return
		}
		if err := os.RemoveAll(dd[0]); err != nil {
			panic(fmt.Sprintf("fx: dirs: reset: %v", err))
		}
	}
}

// CWD changes the current working directory to given directory and
// returns a function which resets this change.  CWD fatales given
// testing instance if given directory is not a temporary directory or
// if the working directory change fails.  It panics if the reset fails.
func CWD(t *testing.T, dir string) (reset func()) {
	if !strings.HasPrefix(dir, os.TempDir()) {
		t.Fatalf("fx: cwd: dir: %s", NoTempDirErr)
	}

	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("fx: cwd: get: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("fx: cwd: %v", err)
	}

	return func() {
		if err := os.Chdir(wd); err != nil {
			panic(fmt.Sprintf("fx: cwd: reset: %v", err))
		}
	}
}

// AddFile adds to given directory a new file with given name and given
// content.  AddFile fatales if given dir is not a temporary directory,
// if the file already exists or os.WriteFile fails.  AddFile panics if
// reset fails.
func AddFile(t *testing.T, dir, name, content string) (reset func()) {

	if !strings.HasPrefix(dir, os.TempDir()) {
		t.Fatalf("fx: add file: dir: %s", NoTempDirErr)
	}

	fl := filepath.Join(dir, name)
	if _, err := os.Stat(fl); err == nil {
		t.Fatalf("fx: add file: already exists")
	}

	if err := os.WriteFile(fl, []byte(content), 0644); err != nil {
		t.Fatalf("fx: add file: write: %v", err)
	}
	return func() {
		if err := os.Remove(fl); err != nil {
			panic(fmt.Sprintf("fx: add file: reset: %v", err))
		}
	}
}

// AddMod adds to given directory a go.mod file with given module name.
// It fatales/panics iff subsequent [fx.AddFile] call fatales/panics.
func AddMod(t *testing.T, dir, module string) (reset func()) {
	return AddFile(t, dir, "go.mod", fmt.Sprintf("module %s", module))
}

// Dir spares in case of several file systems operation the repeating
// providing of testing.T and dir arguments.  It also accumulates reset
// functions and allows to reset them all at once.
type Dir struct {
	T     *testing.T
	Name  string
	reset []func()
}

// NewDir creates a new temp-dir leveraging t.TempDir.
func NewDir(t *testing.T) *Dir {
	d := Dir{T: t}
	d.Name = t.TempDir()
	return &d
}

// MkPath calls [fx.MkPath] using its T and Name property for the
// respective first two arguments.
func (d *Dir) MkPath(dd ...string) (_ *Dir, path string) {
	path, reset := MkPath(d.T, d.Name, dd...)
	d.reset = append(d.reset, reset)
	return d, path
}

// CWD changes the working directory to this directory's name by
// invoking [fx.CWD].
func (d *Dir) CWD() *Dir {
	d.reset = append(d.reset, CWD(d.T, d.Name))
	return d
}

// MkFile adds a file with given name and content to this directory by
// invoking [fx.MkFile].
func (d *Dir) MkFile(name, content string) *Dir {
	d.reset = append(d.reset, AddFile(d.T, d.Name, name, content))
	return d
}

var rePkgComment = regexp.MustCompile(`(?s)^(\s*?\n|// .*?\n|/\*.*\*/)*`)

// MkPkgFile adds a file with given content prefixing its content with a
// package declaration if missing.
func (d *Dir) MkPkgFile(pkg, name, content string) *Dir {
	if !strings.Contains(content, fmt.Sprintf("package %s", pkg)) {
		content = rePkgComment.ReplaceAllString(
			content, fmt.Sprintf("$1\npackage %s\n\n", pkg))
		content = strings.TrimLeft(content, "\n")
	}
	if !strings.HasSuffix(name, ".go") {
		name = fmt.Sprintf("%s.go", name)
	}
	d.reset = append(d.reset, AddFile(
		d.T, d.Name, filepath.Join(pkg, name), content))
	return d
}

// MkPkgFile adds a test file with given content prefixing its content
// with a package declaration and suffixes "_test.go" to the name if
// missing.
func (d *Dir) MkPkgTest(pkg, name, content string) *Dir {
	if !strings.HasSuffix(name, "_test.go") {
		name = fmt.Sprintf("%s%s", name, "_test.go")
	}
	return d.MkPkgFile(pkg, name, content)
}

// MkMod adds to this directory a go.mod file with given module-name by
// invoking [fx.MkMod].
func (d *Dir) MkMod(module string) *Dir {
	d.reset = append(d.reset, AddMod(d.T, d.Name, module))
	return d
}

// Reset calls all collected reset functions in inverse order.
func (d *Dir) Reset() {
	for i := len(d.reset) - 1; i >= 0; i-- {
		d.reset[i]()
	}
	d.reset = []func(){}
}
