package main

import (
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/kubitre/diplom/slaveexecutor/config"
	"github.com/kubitre/diplom/slaveexecutor/core"
	"github.com/kubitre/diplom/slaveexecutor/discovery"
	"github.com/kubitre/diplom/slaveexecutor/routes"
)

func main() {
	conf, err := config.ConfiguringService()
	if err != nil {
		log.Println("can not configuring service: ", err)
		os.Exit(1)
	}

	core, err := core.NewCoreSlaveRunner(conf)
	if err != nil {
		log.Println("can not start core for slave executor", err)
		os.Exit(1)
	}

	slaveRouter := routes.InitNewSlaveRouter(core)
	slaveRouter.ConfigureRouter()
	muxRouter := slaveRouter.GetRouter()
	freePort, errPort := discovery.GetAvailablePort()
	if errPort != nil {
		log.Println("can not get a free port in system: ", errPort)
		os.Exit(1)
	}
	log.Println("initialize api on port: ", strconv.Itoa(freePort))
	if err := http.ListenAndServe(":"+strconv.Itoa(freePort), muxRouter); err != nil {
		log.Println("can not start slave executor: ", err)
		os.Exit(1)
	}
}
