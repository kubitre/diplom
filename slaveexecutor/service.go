package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/kubitre/diplom/slaveexecutor/config"
	"github.com/kubitre/diplom/slaveexecutor/core"
	"github.com/kubitre/diplom/slaveexecutor/routes"
)

func handlingGracefullShutdown(sig chan os.Signal, runCore *core.CoreSlaveRunner) {
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

	runCore.Discovery.UnregisterCurrentService()
	os.Exit(0)
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig)
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
	log.Println("initialize api on port: ", strconv.Itoa(conf.API_PORT))
	go handlingGracefullShutdown(sig, core)
	if err := http.ListenAndServe(":"+strconv.Itoa(conf.API_PORT), muxRouter); err != nil {
		log.Println("can not start slave executor: ", err)
		os.Exit(1)
	}
}
