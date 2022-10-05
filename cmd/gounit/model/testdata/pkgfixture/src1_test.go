// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

/* expected lines of code 3, test code 3, doc 4 */

package pkgfixture // ignore

import "testing" // ignore

func TestSomething(t *testing.T) {
	if true != true {
		t.Error("true is supposed to be true")
	}
}
