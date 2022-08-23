// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package gounit

import (
	"io"
	"io/fs"
	"os"
	fp "path/filepath"
)

type FSfx struct {
	*FS
	mock *fsTMocker
}

func NewFS(t *T) *FSfx {
	return &FSfx{FS: t.FS()}
}

// Mock returns a file system tools mocker to mockup file system
// functions used by Dir and TmpDir instance which can fail.  Note all
// created Dir and TmpDir instances from this FS instance will use the
// mocked filesystem tools including the ones which were created before
// the mocking; use [FSTMocker.Reset] to go back to the defaults if
// necessary.
func (fx *FSfx) Mock() *fsTMocker {
	if fx.mock == nil {
		fx.tools = defaultFSTools.copy()
		fx.mock = &fsTMocker{fs: fx.FS}
	}
	return fx.mock
}

// A fsTMocker allows to set potentially failing file system operations
// which are used by Dir and TmpDir.
type fsTMocker struct{ fs *FS }

// Stat defaults to and has the semantics of os.Stat
func (m *fsTMocker) Stat(f func(string) (fs.FileInfo, error)) {
	m.fs.tools.Stat = f
}

// Mkdir defaults to and has the semantics of os.Mkdir
func (m *fsTMocker) Mkdir(f func(string, fs.FileMode) error) {
	m.fs.tools.Mkdir = f
}

// MkdirAll defaults to and has the semantics of os.MkdirAll
func (m *fsTMocker) MkdirAll(f func(string, fs.FileMode) error) {
	m.fs.tools.MkdirAll = f
}

// Remove defaults to and has the semantics of os.Remove
func (m *fsTMocker) Remove(f func(string) error) {
	m.fs.tools.Remove = f
}

// RemoveAll defaults to and has the semantics of os.RemoveAll
func (m *fsTMocker) RemoveAll(f func(string) error) {
	m.fs.tools.RemoveAll = f
}

// Symlink defaults to and has the semantics of os.Symlink
func (m *fsTMocker) Symlink(f func(string, string) error) {
	m.fs.tools.Symlink = f
}

// Open defaults to and has the semantics of os.Open.
func (m *fsTMocker) Open(f func(string) (*os.File, error)) {
	m.fs.tools.Open = f
}

// Create defaults to and has the semantics of os.Create.
func (m *fsTMocker) Create(f func(string) (*os.File, error)) {
	m.fs.tools.Create = f
}

// ReadDir defaults to and has the semantics of os.ReadDir.
func (m *fsTMocker) ReadDir(f func(string) ([]fs.DirEntry, error)) {
	m.fs.tools.ReadDir = f
}

// ReadFile defaults to and has the semantics of os.ReadFile.
func (m *fsTMocker) ReadFile(f func(string) ([]byte, error)) {
	m.fs.tools.ReadFile = f
}

// WriteFile defaults to and has the semantics of os.WriteFile.
func (m *fsTMocker) WriteFile(f func(string, []byte, fs.FileMode) error) {
	m.fs.tools.WriteFile = f
}

// Chmod defaults to and has the semantics of os.Chmod.
func (m *fsTMocker) Chmod(f func(string, fs.FileMode) error) {
	m.fs.tools.Chmod = f
}

// Copy defaults to and has the semantics of io.Copy
func (m *fsTMocker) Copy(f func(io.Writer, io.Reader) (int64, error)) {
	m.fs.tools.Copy = f
}

// Walk defaults to and has the semantics of filepath.Walk.
func (m *fsTMocker) Walk(f func(string, fp.WalkFunc) error) {
	m.fs.tools.Walk = f
}

// Caller default to and has the semantics of runtime.Caller
func (m *fsTMocker) Caller(f func(int) (uintptr, string, int, bool)) {
	m.fs.tools.Caller = f
}

// Reset resets mocked functions to their default.
func (m *fsTMocker) Reset() { m.fs.tools = defaultFSTools.copy() }
