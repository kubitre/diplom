package core

import (
	"log"
	"time"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/masterexecutor/config"
	"github.com/kubitre/diplom/masterexecutor/discovery"
	"github.com/kubitre/diplom/masterexecutor/routes"
	"github.com/kubitre/diplom/masterexecutor/slaves"
)

/*RunnerCore - ядро master ноды*/
type RunnerCore struct {
	RouterRunner  *routes.RunnerRouter
	Discovery     *discovery.Discovery
	SlaveMoniring *slaves.SlaveMonitoring
}

/*InitNewCore - инициализация ядра текущего сервиса*/
func InitNewCore(config *config.ConfigurationRunner) (*RunnerCore, error) {
	slaveMonitor, err := slaves.InitializeNewSlaveMonitoring(config.MaxTaskPerSlave)
	if err != nil {
		return nil, err
	}
	slaveMonitor.LastUsingService <- slaves.INIT_USED_SLAVE
	return &RunnerCore{
		SlaveMoniring: slaveMonitor,
		RouterRunner:  routes.InitializeRunnerRouter(slaveMonitor, config),
		Discovery:     discovery.InitializeDiscovery(config),
	}, nil
}

/*Run - запуск роутера, discovery, получение информации о слейвах*/
func (core *RunnerCore) Run(config *config.ConfigurationRunner) {
	core.Discovery.NewClientForConsule()
	core.Discovery.RegisterServiceWithConsul()
	core.RouterRunner.ConfiguringRoutes()
	go core.checkerNewSlave()

}

func (core *RunnerCore) checkerNewSlave() {
	for {
		log.Println("start finding slaves")
		foundedSlaves := core.Discovery.CheckNewSlaves()
		log.Println("founded services: ", foundedSlaves)
		core.SlaveMoniring.CompareAndSave(foundedSlaves)
		time.Sleep(time.Second * 15)
	}
}

/*UnregisterService - де регистрация сервиса из consul*/
func (core *RunnerCore) UnregisterService() {
	core.Discovery.UnregisterCurrentService()
}

/*GetRouter - получение сконфигурированного роутера*/
func (core *RunnerCore) GetRouter() *mux.Router {
	return core.RouterRunner.GetRouterMux()
}
