// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

var fxExp = map[string][]string{
	// go/pass package should display all test and sub-test names
	"go/pass": {"test pass 1", "test pass 2", "test pass 3",
		"test pass 4", "p4 sub 1", "p4 sub 2", "p4 sub 3", "p4 sub 4",
		"test pass 5", "p5 sub 1", "p5 sub 2", "p5 sub 3", "p5 sub 4",
	},
	// go/pass folded should display test-names only (see fxNotExp)
	"go/pass: folded": {"test pass 1", "test pass 2", "test pass 3",
		"test pass 4", "test pass 5"},
	// mixed/pass should contain initially "go-tests" and the suites
	"mixed/pass": {"mixed/pass", "go-tests", "suite 1", "suite 2",
		"suite 3", "suite 4", "suite 5"},
	"mixed/pass go folded subs": {"mixed/pass", "go-tests", "test pass 1",
		"test pass 2", "test pass 3", "test pass 4", "test pass 5"},
	"mixed/pass first suite": {"mixed/pass", "suite 1",
		"suite test 1 1", "suite test 1 2", "suite test 1 3",
		"suite test 1 4", "suite test 1 5"},
	"mixed/pass second suite": {"mixed/pass", "suite 2",
		"suite test 2 1", "suite test 2 2", "suite test 2 3",
		"suite test 2 4", "suite test 2 5"},
	"mixed/pass go unfold": {"mixed/pass", "go-tests", "test pass 4",
		"p4 sub 1", "p4 sub 2", "p4 sub 3", "p4 sub 4"},
	"mixed/pp/pkg0": {"mixed/pp/pkg0"},
	"mixed/pp": {"pkg0", "pkg1", "pkg2", "pkg3", "pkg4", "pkg5", "pkg6",
		"pkg7", "pkg8", "pkg9"},
	"mixed/pp/pkg3": {"mixed/pp/pkg3"},
}

var fxNotExp = map[string][]string{
	// folded shouldn't contain the sub-tests of the go/pass-package
	"go/pass: folded": {"p4_sub_1", "p4_sub_2", "p4_sub_3", "p4_sub_4"},
	// mixed/pass shouldn't contain initially sub-tests or suite-test
	"mixed/pass": {"Suite_test", "_sub_"},
	// mixed/pass go folded subs shouldn't contain suites and subs
	"mixed/pass go folded subs": {"Suite_", "_sub_"},
	"mixed/pass first suite": {"go-tests", "Suite_2", "Suite_3", "Suite_4",
		"Suite_5"},
	"mixed/pass second suite": {"go-tests", "Suite_1", "Suite_3", "Suite_4",
		"Suite_5"},
	"mixed/pass go unfold": {"TestPass_1", "TestPass_2", "TestPass_3",
		"TestPass_5"},
}
