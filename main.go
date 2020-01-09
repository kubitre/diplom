package main

import (
	"log"

	"github.com/kubitre/diplom/gitmod"
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
	log.Println("Tasks: ", runn.Tasks)

	gt := gitmod.Git{}
	if err = gt.CloneRepo("https://github.com/kubitre/for_diplom.git"); err != nil {
		log.Panic(err)
	}

}
