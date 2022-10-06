// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

var fxExp = map[string][]string{
	// go/pass should display test-names only (see fxNotExp)
	"go/pass": {"go/pass", "pass 1", "pass 2", "pass 3",
		"pass 4", "pass 5"},
	"go/pass suite": {"go/pass", "pass 4", "p4 sub 1", "p4 sub 2",
		"p4 sub 3", "p4 sub 4"},
	"mixed/pass init": {"mixed/pass", "suite 3", "suite test 3 1",
		"suite test 3 2", "suite test 3 3", "suite test 3 4",
		"suite test 3 5"},
	"mixed/pass fold suite": {"mixed/pass", "go-tests", "suite 1",
		"suite 2", "suite 3", "suite 4", "suite 5"},
	"mixed/pass unfold suite": {"mixed/pass", "suite 2",
		"suite test 2 1", "suite test 2 2", "suite test 2 3",
		"suite test 2 4", "suite test 2 5"},
	"mixed/pass go folded subs": {"mixed/pass", "go-tests", "pass 1",
		"pass 2", "pass 3", "pass 4", "pass 5"},
	"mixed/pass go unfolded suite": {"mixed/pass", "go-tests", "pass 4",
		"p4 sub 1", "p4 sub 2", "p4 sub 3", "p4 sub 4"},
	"mixed/pass first suite": {"mixed/pass", "suite 1",
		"suite test 1 1", "suite test 1 2", "suite test 1 3",
		"suite test 1 4", "suite test 1 5"},
	"mixed/pass second suite": {"mixed/pass", "suite 2",
		"suite test 2 1", "suite test 2 2", "suite test 2 3",
		"suite test 2 4", "suite test 2 5"},
	"mixed/pass go unfold": {"mixed/pass", "go-tests", "pass 4",
		"p4 sub 1", "p4 sub 2", "p4 sub 3", "p4 sub 4"},
	"mixed/pp/pkg0": {"mixed/pp/pkg0"},
	"mixed/pp": {"pkg0", "pkg1", "pkg2", "pkg3", "pkg4", "pkg5", "pkg6",
		"pkg7", "pkg8", "pkg9"},
	"mixed/pp/pkg3": {"mixed/pp/pkg3"},
	"logging go-tests": {"logging", "go-tests", "go log",
		"go-test log", "go suite log"},
	"logging go-suite": {"logging", "go-tests", "go suite log",
		"go sub log", "go sub-test log"},
	"logging suite": {"logging", "suite", "init-log:", "suite init log",
		"suite test log", "suite-test log", "finalize-log",
		"suite finalize log"},
	"logging folded": {"logging", "go-tests", "suite"},
	"fail compile": {"fail/compile", "shell exit error: exit status 2:",
		"fail/compile/compile_test.go:7:33:", "undefined: Sum",
		"FAIL example.com/gounit/controller/golden/fail/compile "},
	"fail mixed go-suite": {"test2 fail", "test4 failing", "p4 sub 2"},
	"fail mixed suite": {
		"go-tests", "suite 2", "suite 4", "suite test 4 3"},
	"fail pp": {"fail/pp/fail1", "fail/pp/fail2"},
	"fail pp collapsed": {
		"fail/pp/fail1", "fail/pp/fail2", "fail/pp/pass"},
	"del before": {"del/pkg1", "del/pkg2"},
	"panic": {
		"panic/pkg1", "panic: runtime error: index out of range"},
}

var fxNotExp = map[string][]string{
	// go-pass shouldn't contain the sub-tests of the go/pass-package
	"go/pass": {"p4_sub_1", "p4_sub_2", "p4_sub_3", "p4_sub_4"},
	"go/pass init": {"go-tests", "suite 1", "suite 2", "suite 4",
		"suite 5"},
	"go/pass suite": {"pass 1", "pass 2", "pass 3",
		"pass 4", "pass 5"},
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
