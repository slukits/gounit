/*
Gounit watches package directories of a go module and reports test
results on source file changes.  It watches always the current working
directory and its descending packages.  It fails to do so if the current
directory is not inside an module.

Usage:

	gounit

By default gounit reports of watched sources always the latest test
results associated with latest modifications.  A user can request to get
all packages of a watch sources directory listed.  While a requested
suite view shows all test-suites of the "current" package.  Finally in
the settings a user can choose to switch on/off: race, vet and stats.
The later reports about a package and the watched sources how many
source files are there and how many of them are for testing; it reports
how many lines of code, how many of them are for testing and how many
lines of documentation are found in a package and overall in the watched
sources.  If stats are on gounit also reports a diff of the first stats
calculations and the last when it is quit by the user.  Sample ui:

	github.com/slukits/gounit: cmd/gounit


	view: 21/0 243ms; 10/5 files, 1454/684 340 lines

	A new view displays initially given                          16ms
	    message                                                   2ms
	    status                                                    4ms
	    main_info                                                 3ms
	    buttons                                                   7ms

	A view 17/0 53ms
	...


	pkgs/suites: p/s; tests: t/f; files: s/t, lines: c/t d

	 [p]kgs  [s]uites  [v]et=off   [r]ace=off   [s]tats=on   [q]uit

The message-bare displays the module and its watched sources-directory.

The reporting component in this case shows watched testing package
"view" with its total number of tests and failed tests.  The total
number of source files and the number of test files.  The total number
of source code lines and the lines of test code.  And finally the number
of documentation lines.  The "current" package line is followed by the
"current" suite line which is indented followed by the suite tests.  The
rest of the screen til the status bar is filled with the view package's
remaining test-suites.

The statusbar reports the number of packages in the watched sources, the
total number of tests in the watched directory and the number of failed
tests; the total number of code lines in the watched directory, the
lines of test-code and the lines of documentation in the package.

Finally we have a button bar whose buttons may be clicked or executed by
pressing associated key.
*/
package main

import (
	"github.com/slukits/gounit/cmd/gounit/controller"
)

func main() {
	controller.New(controller.InitFactories{})
}
