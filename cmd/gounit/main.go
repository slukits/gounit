package main

import (
	"log"

	"github.com/slukits/gounit/cmd/gounit/controller"
	"github.com/slukits/gounit/pkg/module"
	"github.com/slukits/lines"
)

func main() {
	controller.New(log.Fatal, &module.Sources{}, lines.New)
}
