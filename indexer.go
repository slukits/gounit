// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// indexer provides the suiteTestsIndexer-type whose only task it is to
// index the suite methods of a test-file by their appearance.

package gounit

import (
	"go/ast"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
)

// type suiteTestsIndexer map[string]map[string]map[string]int

var indexer = suiteTestsIndexer{}

// suiteTestsIndexer provieds *ensureIndexingOf(testFileName)* which
// parses a test-file's test-suites and their tests to index the later
// in order of their appearance.  While get(testFileName, suiteName,
// testName) retrieves the mapping created by ensureIndexingOf.  These
// operations are concurrency save, e.g. while a file is parsed no index
// may be retrieved and vice versa.  Since each suite-runner of a given
// test-file calls ensureIndexingOf before running any test it is
// guaranteed that a test-file's suit-test indices are calculated before
// any of them is retrieved.
type suiteTestsIndexer struct {
	// In case suites of the same package are run in parallel one of the
	// 'runners' per test file indexes the suite-methods.  If an parallel
	// running suite of the same test-file wants to access the indexer it
	// should be ensured the indices calculation is finished.
	mutex    sync.Mutex
	_Indexer map[string]map[string]map[string]int
}

// get returns a given test's index which is defined in given suite
// which in turn is defined in given test-file.
func (i *suiteTestsIndexer) get(file, suite, test string) int {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	return i._Indexer[file][suite][test]
}

// ensureIndexingOf parses given ast-file (with given name) for its
// suites and suite-tests whereas the later are mapped to indices in
// order of appearance.
func (i *suiteTestsIndexer) ensureIndexingOf(f *ast.File, name string) {
	i.mutex.Lock()
	defer i.mutex.Unlock()

	if _, ok := i._Indexer[name]; ok {
		return
	}
	if i._Indexer == nil {
		//             file-name suite-name test-name  index
		i._Indexer = map[string]map[string]map[string]int{}
	}
	i._Indexer[name] = map[string]map[string]int{}
	i._ParseSuites(f, name)
	if len(i._Indexer[name]) == 0 {
		return
	}
	i._ParseSuiteTests(f, name)
}

func (i *suiteTestsIndexer) _Has(suite, inFile string) bool {
	_, ok := i._Indexer[inFile][suite]
	return ok
}

// _IsIdent helps investigating if a function's receiver field type
// refers to a known test-suite by returning given field-type's
// identifier-name if their is any.
func (i *suiteTestsIndexer) _IsIdent(fldType ast.Expr) (string, bool) {
	if ident, ok := fldType.(*ast.Ident); ok {
		return ident.Name, true
	}

	starExpr, ok := fldType.(*ast.StarExpr)
	if !ok {
		return "", false
	}
	ident, ok := starExpr.X.(*ast.Ident)
	if !ok {
		return "", false
	}

	return ident.Name, true
}

var isLower = regexp.MustCompile(`^[_a-z]`)

// _IsSuiteTest returns a suite's name, the suite-test's name and true in
// case given function declaration represents a suite-test; zero-strings
// and false otherwise.
func (i *suiteTestsIndexer) _IsSuiteTest(fd *ast.FuncDecl, fl string) (
	suite, test string, _ bool) {

	if fd.Recv == nil {
		return "", "", false
	}
	if strings.Contains(special, fd.Name.Name) ||
		len(fd.Type.Params.List) != 1 ||
		isLower.MatchString(fd.Name.Name) {
		return "", "", false
	}
	for _, field := range fd.Recv.List {
		name, ok := i._IsIdent(field.Type)
		if !ok {
			continue
		}
		if !i._Has(name, fl) {
			continue
		}
		return name, fd.Name.Name, true
	}
	return "", "", false
}

func (i *suiteTestsIndexer) _ParseSuiteTests(f *ast.File, fname string) {
	ast.Inspect(f, func(n ast.Node) bool {
		switch n := n.(type) {
		case *ast.FuncDecl:
			if suite, name, ok := i._IsSuiteTest(n, fname); ok {
				idx := len(i._Indexer[fname][suite])
				i._Indexer[fname][suite][name] = idx
				return true
			}
		}
		return true
	})
}

const gounitPath = `"github.com/slukits/gounit"`
const suitePath = gounitPath

// _ParseImports determines if the Suite type is imported into a
// test-file.  E.g. it is a necessary for a test-file defining test
// suites.
func (i *suiteTestsIndexer) _ParseImports(f *ast.File) (string, bool) {
	for _, imp := range f.Imports {
		if imp.Path.Value == gounitPath || imp.Path.Value == suitePath {
			if imp.Name != nil {
				if imp.Name.Name != "." {
					return imp.Name.Name, true
				}
			} else {
				return filepath.Base(strings.Trim(
					imp.Path.Value, `"`)), true
			}
			return "", true
		}
	}
	return "", false
}

// _IsSuite returns true if given struct-type embeds the Suite-struct.
func (i *suiteTestsIndexer) _IsSuite(
	st *ast.StructType, suiteslc string) bool {

	for _, field := range st.Fields.List {
		if suiteslc == "" {
			if ident, ok := field.Type.(*ast.Ident); ok {
				if ident.Name == "Suite" {
					return true
				}
			}
		} else {
			if slc, ok := field.Type.(*ast.SelectorExpr); ok {
				if slc.Sel.Name != "Suite" {
					return false
				}
				if selIdent, ok := slc.X.(*ast.Ident); ok {
					if selIdent.Name == suiteslc {
						return true
					}
				}
			}
		}
	}
	return false
}

// _IsStruct returns a struct's name its ast-representation and true in
// case given node is a struct-definition; zeros and false otherwise.
func (i *suiteTestsIndexer) _IsStruct(n ast.Node) (
	string, *ast.StructType, bool) {

	var typeSpec *ast.TypeSpec
	var structType *ast.StructType
	var ok bool
	if typeSpec, ok = n.(*ast.TypeSpec); !ok {
		return "", nil, false
	}
	if structType, ok = typeSpec.Type.(*ast.StructType); !ok {
		return "", nil, false
	}
	return typeSpec.Name.Name, structType, true
}

// _ParseSuites extracts all the struct types from a test-file embedding
// the Suite-type.  NOTE since suite-methods may be defined 'before' the
// suite-type is defined the suites are parsed in an extra first pass.
func (i *suiteTestsIndexer) _ParseSuites(f *ast.File, fname string) {
	slc, ok := i._ParseImports(f)
	if !ok {
		return
	}

	ast.Inspect(f, func(n ast.Node) bool {
		name, structType, ok := i._IsStruct(n)
		if !ok {
			return true
		}
		if !i._IsSuite(structType, slc) {
			return true
		}
		i._Indexer[fname][name] = map[string]int{}
		return true
	})
}
