// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

var fxExp = map[string][]string{
	// go/pass package should display all test and sub-test names
	"go/pass": {"TestPass_1", "TestPass_2", "TestPass_3",
		"TestPass_4", "p4_sub_1", "p4_sub_2", "p4_sub_3", "p4_sub_4",
		"TestPass_5", "p5_sub_1", "p5_sub_2", "p5_sub_3", "p5_sub_4",
	},
	// go/pass folded should display test-names only (see fxNotExp)
	"go/pass: folded": {"TestPass_1", "TestPass_2", "TestPass_3",
		"TestPass_4", "TestPass_5"},
	// mixed/pass should contain initially "go-tests" and the suites
	"mixed/pass": {"mixed/pass", "go-tests", "Suite_1", "Suite_2",
		"Suite_3", "Suite_4", "Suite_5"},
	"mixed/pass go folded subs": {"mixed/pass", "go-tests", "TestPass_1",
		"TestPass_2", "TestPass_3", "TestPass_4", "TestPass_5"},
	"mixed/pass first suite": {"mixed/pass", "Suite_1",
		"Suite_test_1_1", "Suite_test_1_2", "Suite_test_1_3",
		"Suite_test_1_4", "Suite_test_1_5"},
	"mixed/pass second suite": {"mixed/pass", "Suite_2",
		"Suite_test_2_1", "Suite_test_2_2", "Suite_test_2_3",
		"Suite_test_2_4", "Suite_test_2_5"},
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
}
