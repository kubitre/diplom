package main

import (
	"log"

	"github.com/kubitre/diplom/slaveexecutor/config"
)

func main() {
	conf, err := config.ConfiguringService()
	if err != nil {
		log.Panic("can not configuring service: ", err)
	}
	
}