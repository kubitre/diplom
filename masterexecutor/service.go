package main

import (
	"log"
	"net/http"
	"strconv"

	"github.com/kubitre/diplom/masterexecutor/config"
	"github.com/kubitre/diplom/masterexecutor/routes"
)

func main() {
	config, err := config.ConfiureRunnerMaster()
	if err != nil {
		log.Panic("can not configuring service by config: ", err)
	}
	runnerRouter := routes.InitializeRunnerRouter()
	runnerRouter.ConfiguringRoutes()
	router := runnerRouter.GetRouterMux()
	if err := http.ListenAndServe(":"+strconv.Itoa(config.APIPort), router); err != nil {
		log.Panic("can not start listen and serve: ", err)
	}
}
