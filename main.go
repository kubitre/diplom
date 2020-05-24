package main

import (
	"os"

	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/core"
	log "github.com/sirupsen/logrus"
)

func moduleCanBeStart() (serviceConfig *config.ServiceConfig) {
	serviceConfig, err := config.ConfigureService()
	if err != nil {
		log.Warn(err)
	}
	return serviceConfig
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

	}
	// loadedFile, err := tools.TestLoadFile("./specifications/task.yaml")
	// if err != nil {
	// 	log.Panic(err)
	// }
	// runn, err := tools.ParseObj(loadedFile)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// log.Println(runn)
	// runner, err := core.NewCoreSlaveRunner(nil, nil)
	// if err != nil {
	// 	log.Panic(err)
	// }
	// if err := runner.CreatePipeline(runn); err != nil {
	// 	log.Panic(err)
	// }

}
