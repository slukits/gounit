// copyright (c) 2022 stephan lukits. all rights reserved.
// use of this source code is governed by a mit-style
// license that can be found in the license file.

package controller

import (
	"strings"

	"github.com/slukits/gounit/cmd/gounit/view"
)

func viewHelp() *report {
	return &report{
		flags: view.RpClearing,
		ll:    strings.Split(strings.TrimSpace(help), "\n"),
	}
}

const help = `
Press 'q' or 'ctrl+d' or 'ctrl+c' key to quit this application.

Right click or space-key-press scrolls down in circling fashion.

Page up/down scrolls up and down.

The typical use case of gounit is to start it in the package you are
working on.  Write a test in that package (or one of its sub-packages)
which fails.  gounit will automatically report the suite/go-tests with
the failing test.  Make the test "green", refactor and write again a
failing test etc.

The first line initially displays the module name and the watched
source directory.

Below is initially the most recently modified package reported.  That 
is if all tests are passing.  If there is a failing test then always 
the failing test of the most recently modified test file is reported.

NOTE if a package contains only go-test, i.e. not gounit-suites, the
go tests are reported sorted by name and after them go-suites, i.e. 
go-tests with sub-tests, are reported folded.  If a package contains
also test-suites the go-tests and all suites of the package are
reported folded.

Clicking on a reported package name folds that package and shows all
packages in the watched source directory.  Clicking in the packages
view on a package name will report this package's tests.

Clicking on a reported suite's name folds this suite and shows the
reported package's go tests and test-suites.  Clicking on "go-tests"
or a suite name will then report the go tests or the clicked suite.

NOTE selectable lines like package names, go-tests or suites can be
selected using the j or k key for highlighting the next/previous
selectable line.  Pressing enter on a highlighted line will select
it respectively fold/unfold the selected package/suite accordingly.

Buttons in the bottom may be selected by clicking on them or pressing
the embraced key.

[v]et switches the Go vet execution for test-runs on and off.  I.e.
       on [v]et=off the "go test" command is run with the "-vet=off"
       flag.  Otherwise this flag is omitted.

[r]ace switches the Go's race detector for test-runs on and off.  I.e. 
       if [r]ace=on the "go test" command is run with the "-race"-flag.
       Otherwise this flag is omitted.

NOTE if vet or race is switched on while a particular package is
reported this package's tests are rerun with the according flags set.
If you are in the packages-view while you switch, the switch takes 
effect at the next package source-change.

[s]tats switches reporting source statistics on and off.  I.e. if
       [s]tats=on for each package is the number of source files
       including test-files, the number of test-files, the number
       of source code lines including test-code, the number of test-code
       lines and the number of documentation lines reported.  In the
       status bar the same information for the total of the watched 
       source directory is provided.

[m]ore makes more buttons available like [h]elp showing this help text,
       [a]bout with copyright and license information.

Happy coding!
`
