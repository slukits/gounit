// Copyright (c) 2022 Stephan Lukits. All rights reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package controller

import (
	"fmt"
	"os"
	"strings"

	"github.com/slukits/lines"
)

const (
	NotImplemented = "gounit: package support not implemented yet"
	NotSupported   = "gounit: argument %s: is not supported; only -p"
)

// New starts the application and blocks until a quit event occurs.
// Fatale errors are reported to ftl while ll is used to initialize the
// ui and start the event loop.
func New(
	ftl func(...interface{}),
	ll func(lines.Componenter) *lines.Events,
) {

	if len(os.Args) > 1 {
		if strings.HasSuffix(os.Args[1], "p") && strings.Contains(
			"--p", os.Args[1],
		) { // TODO: implement
			ftl(NotImplemented)
			return
		}
		ftl(fmt.Sprintf(NotSupported, os.Args[1]))
		return
	}
	ll(nil).Listen()
}
