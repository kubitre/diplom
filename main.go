package main

import (
	"log"

	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/tools"
)

func main() {
	loadedFile, err := tools.TestLoadFile("./specifications/task.yaml")
	if err != nil {
		log.Panic(err)
	}
	runn, err := tools.ParseObj(loadedFile)
	if err != nil {
		log.Panic(err)
	}
	runner, err := core.NewCoreSlaveRunner(nil, nil)
	if err != nil {
		log.Panic(err)
	}
	if err := runner.CreatePipeline(runn); err != nil {
		log.Panic(err)
	}

}
