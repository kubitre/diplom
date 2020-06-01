package main

import (
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/routes"
	"github.com/kubitre/diplom/routes/route_default"
	"github.com/kubitre/diplom/routes/route_portal"
	"github.com/kubitre/diplom/services"
	log "github.com/sirupsen/logrus"
)

func moduleCanBeStart() (serviceConfig *config.ServiceConfig) {
	serviceConfig, err := config.ConfigureService()
	if err != nil {
		log.Warn(err)
	}
	return serviceConfig
}

func runRouter(router *mux.Router, serviceConfig *config.ServiceConfig) {
	if err := http.ListenAndServe(":"+strconv.Itoa(serviceConfig.APIPORT), router); err != nil {
		log.Panic("can not be starting service: ", err)
	}
}

func initMasterRunnerRouterByPlugin(
	masterService *services.MasterRunnerService,
	serviceConfig *config.ServiceConfig) routes.IMaster {
	switch serviceConfig.ServicePlugin {
	case config.PLUGINPORTAL:
		log.Info("runner will start with portal plugin")
		router := route_portal.InitializeMasterRunnerRouter(masterService)
		return router
	default:
		log.Info("runner will start with default plugin")
		router := route_default.InitializeMasterRunnerRouter(masterService)
		return router
	}
}
func handlingGracefullShutdown(sig chan os.Signal, masterCore *core.MasterRunnerCore, slaveCore *core.SlaveRunnerCore) {
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
	if masterCore != nil {
		masterCore.UnregisterService()
	}
	if slaveCore != nil {
		slaveCore.UnregisterService()
	}
	os.Exit(0)
}

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig)
	serviceConfig := moduleCanBeStart()
	log.SetLevel(log.DebugLevel)
	switch serviceConfig.ServiceType {
	case config.SERVICESLAVE:
		runnerConfig, errConfiguring := config.ConfigureRunnerSlave()
		if errConfiguring != nil {
			log.Warn("can not correct configuring: ", errConfiguring)
		}
		runner, err := core.NewCoreSlaveRunner(runnerConfig, serviceConfig)
		if err != nil {
			log.Error("slave service can not be start: ", err)
			os.Exit(1)
		}
		runner.RunWorkers()
		routerSlave := routes.InitNewSlaveRunnerRouter(runner)
		routerSlave.ConfigureRouter()
		log.Info("start agent as slave")
		go handlingGracefullShutdown(sig, nil, runner)
		runRouter(routerSlave.GetRouter(), serviceConfig)
	default:
		runnerConfig, errConfiguring := config.ConfiureRunnerMaster()
		if errConfiguring != nil {
			log.Warn("can not correct configuring: ", errConfiguring)
		}
		masterService, errService := services.InitializeMasterRunnerService(serviceConfig, runnerConfig)
		if errService != nil {
			log.Error("can not initialize master runner service: ", errService)
			os.Exit(1)
		}
		routerMaster := initMasterRunnerRouterByPlugin(masterService, serviceConfig)
		routerMaster.ConfigureRouter()

		go handlingGracefullShutdown(sig, masterService.GetCore(), nil)
		log.Info("start agent as master")
		runRouter(routerMaster.GetRouter(), serviceConfig)
	}
}
