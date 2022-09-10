// copyright (c) 2022 stephan lukits. all rights reserved.
// use of this source code is governed by a mit-style
// license that can be found in the license file.

package controller

import "strings"

func viewHelp(updVW func(interface{})) {
	h := strings.Split(strings.TrimSpace(help), "\n")
	updVW(&liner{clearing: true, ll: h})
}

const help = `
Press 'q' or 'ctrl+d' or 'ctrl+c' key to quit this application.

A right click or space-key-press scrolls down in circling fashion.
Page up/down scrolls up and down.  Clicking on a suite or test name
the module relative file position of that suite or test is shown for
IDE-terminal implementations supporting to jump to the file position.

gounit has three main-views: the "default-view", the "packages-view" and
the "error-view".  The default-view always reports the most recently
changed package's tests, test suites and their test results.  The 
packages gounit reports about is the current working directory and 
its descendants.  The packages-view provides an overview of all 
currently watched packages.  Both views come also with a suites variant.
The default-view's suites variants doesn't report the test results of
the latest changed suite but a list of all suites in the currently 
reported package.  The packages-view shows additional to all packages 
also the suites of the currently selected package.  Clicking a suite 
turns the suites variant of and makes the selected suite the currently
reported suite.  You can use the buttons at the bottom of the screen to
switch between views.

The error-view is shown if at least one test of the watched packages
failed reporting the failed test(s).  The error-view comes with an
additional error button at the bottom to switch between the other
views and the error view.

Default buttons

[p]kgs switches to the packages-view and back.

[s]uites switches the suites-variant on and off.

[a]rgs lets you define how test-runs are execute and their results are
       reported by providing the switches [v]et, [r]ace and [s]tats.

[m]ore switches to the buttons set: [h]elp, [a]bout and [q]uit.

Args buttons

[v]et turns the vet switch for the next test run on/off.  The default
      is off.

[r]ace turns the race switch for the next test run on/off.  The default
       is off.

[s]tats turns statistics on and off.  This switch extends the reported
        numbers by the number of code lines, how many of them are for
        testing and how many lines of documentation is found.  Stats
        also stores the differences between starting and stopping
        stats for later analysis.
`
