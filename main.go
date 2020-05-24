package main

import (
	"net/http"
	"os"
	"plugin"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/routes"
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
	if err := http.ListenAndServe(""+strconv.Itoa(serviceConfig.APIPORT), router); err != nil {
		log.Panic("can not be starting service: ", err)
	}
}

func initMasterRunnerRouterByPlugin(
	runnerCore *core.MasterRunnerCore,
	serviceConfig *config.ServiceConfig,
	runnerConfig *config.ConfigurationMasterRunner) routes.IMaster {
	module := ""
	switch serviceConfig.ServicePlugin {
	case config.PLUGINPORTAL:
		module = "./routes/route_portal/router_master.so"
	default:
		module = "./routes/route_default/router_master.so"
	}
	plug, err := plugin.Open(module)
	if err != nil {
		log.Error("can not be initialize by plugin: ", module)
		os.Exit(1)
	}
	symLinkRouter, err := plug.Lookup("MasterRouter")
	if err != nil {
		log.Error("can not find router object: ", err)
		os.Exit(1)
	}
	runnerRouter, errAssert := symLinkRouter.(routes.IMaster)
	if !errAssert {
		log.Error("unexpected type from module sym")
		os.Exit(1)
	}
	return runnerRouter
}

func main() {
	serviceConfig := moduleCanBeStart()
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
		routerSlave := routes.InitNewSlaveRunnerRouter(runner)
		routerSlave.ConfigureRouter()

		runRouter(routerSlave.GetRouter(), serviceConfig)
	default:
		runnerConfig, errConfiguring := config.ConfiureRunnerMaster()
		if errConfiguring != nil {
			log.Warn("can not correct configuring: ", errConfiguring)
		}
		runner, err := core.InitNewMasterRunnerCore(runnerConfig, serviceConfig)
		if err != nil {
			log.Error("master service can not be start: ", err)
		}
		routerMaster := initMasterRunnerRouterByPlugin(runner, serviceConfig, runnerConfig)
		routerMaster.ConfigureRouter()

		runRouter(routerMaster.GetRouter(), serviceConfig)
	}
}
