// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

var fxExp = map[string][]string{
	"go/pass": {"TestPass_1", "TestPass_2", "TestPass_3",
		"TestPass_4", "p4_sub_1", "p4_sub_2", "p4_sub_3", "p4_sub_4",
		"TestPass_5", "p5_sub_1", "p5_sub_2", "p5_sub_3", "p5_sub_4",
	},
	"go/pass: folded": {"TestPass_1", "TestPass_2", "TestPass_3",
		"TestPass_4", "TestPass_5"},
}
