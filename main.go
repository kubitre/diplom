package main

import (
	"log"

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
	log.Println(runn)
	// runner, err := core.NewCoreSlaveRunner(nil, nil)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// if err := runner.CreatePipeline(runn); err != nil {
	// 	log.Panic(err)
	// }

}
