// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

// expected lines of code 6, test-code 0, doc 4

package srcstats // ignore

import ( // ignore imports
	"fmt"
	"time"
)

func calculateAnswer(s []string) int {
	return len(s) * 21
}

func TheAnswer() int {
	return calculateAnswer([]string{
		fmt.Sprintf("%v", time.Now()),
		"two",
	}) // ignore
} /* ignore */
