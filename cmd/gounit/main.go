/*
Gounit watches package directories of a go module and reports test
results on source file changes.  It watches always the current working
directory and its nested packages.  It fails to do so if the current
directory is not inside an module.

Usage:

	gounit
*/
package main

import (
	"github.com/slukits/gounit/cmd/gounit/controller"
)

func main() {
	controller.New(nil)
}
