// Copyright (c) 2022 Stephan Lukits. All rights reserved.
//  Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package fx

import (
	"fmt"
	"strings"
)

// SuiteTestsIndexerToString stringifies an indexers reflection map.
func SuiteTestsIndexerToString(
	indices map[string]map[string]map[string]int,
) string {
	lines := []string{}
	for k1 := range indices {
		for k2 := range indices[k1] {
			lines = append(lines, fmt.Sprintf("%s.%s", k1, k2))
			for k3, v := range indices[k1][k2] {
				lines = append(lines, fmt.Sprintf("    %s: %d", k3, v))
			}
		}
	}
	return strings.Join(lines, "\n")
}
