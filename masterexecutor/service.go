package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kubitre/diplom/masterexecutor/config"
	"github.com/kubitre/diplom/masterexecutor/core"
)

func handlingGracefullShutdown(sig chan os.Signal, runCore *core.RunnerCore) {
	for {
		sg := <-sig
		switch sg {
		case syscall.SIGINT, syscall.SIGTERM:
			break
		default:
			continue
		}

		log.Println("init kill: ", sg)
		signal.Reset(sg)
		break
	}

	log.Println("gracefull shutdown")

	runCore.UnregisterService()
	os.Exit(0)
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig)
	config, err := config.ConfiureRunnerMaster()
	if err != nil {
		log.Panic("can not configuring service by config: ", err)
	}
	log.Println("Configuration for current service: ", config)
	runnerCore, err := core.InitNewCore(config)
	if err != nil {
		log.Panic("can not initialize core: ", err)
	}
	runnerCore.Run(config)
	router := runnerCore.GetRouter()
	runnerCore.Discovery.GetService("master-executor", "executor")
	go handlingGracefullShutdown(sig, runnerCore)
	if err := http.ListenAndServe(":"+strconv.Itoa(config.APIPort), router); err != nil {
		log.Panic("can not start listen and serve: ", err)
	}
}
