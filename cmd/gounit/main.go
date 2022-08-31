package main

import (
	"log"

	"github.com/slukits/gounit/cmd/gounit/controller"
	"github.com/slukits/gounit/cmd/gounit/model"
	"github.com/slukits/lines"
)

func main() {
	controller.New(log.Fatal, &model.Sources{}, lines.New)
}
