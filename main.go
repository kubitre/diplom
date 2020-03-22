package main

import (
	"log"

	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/parser"
)

func main() {
	loadedFile, err := parser.TestLoadFile("./specifications/task.yaml")
	if err != nil {
		log.Panic(err)
	}
	runn, err := parser.ParseObj(loadedFile)
	if err != nil {
		log.Panic(err)
	}
	runner, err := core.NewCoreRunner(1, nil)
	if err != nil {
		log.Panic(err)
	}
	if err := runner.CreatePipeline(runn); err != nil {
		log.Panic(err)
	}

}
