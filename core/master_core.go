package core

import (
	"log"
	"time"

	"github.com/gorilla/mux"
)

/*MasterRunnerCore - ядро master ноды*/
type MasterRunnerCore struct {
	RouterRunner  *routes.MasterRunnerRouter
	Discovery     *discovery.Discovery
	SlaveMoniring *slaves.SlaveMonitoring
}

/*InitNewCore - инициализация ядра текущего сервиса*/
func InitNewCore(config *config.ConfigurationRunner) (*MasterRunnerCore, error) {
	slaveMonitor, err := slaves.InitializeNewSlaveMonitoring(config.MaxTaskPerSlave)
	if err != nil {
		return nil, err
	}
	slaveMonitor.LastUsingService <- slaves.INIT_USED_SLAVE
	return &MasterRunnerCore{
		SlaveMoniring: slaveMonitor,
		RouterRunner:  routes.InitializeRunnerRouter(slaveMonitor, config),
		Discovery:     discovery.InitializeDiscovery(config),
	}, nil
}

/*Run - запуск роутера, discovery, получение информации о слейвах*/
func (core *MasterRunnerCore) Run(config *config.ConfigurationRunner) {
	core.Discovery.NewClientForConsule()
	core.Discovery.RegisterServiceWithConsul()
	core.RouterRunner.ConfiguringRoutes()
	go core.checkerNewSlave()

}

func (core *MasterRunnerCore) checkerNewSlave() {
	for {
		log.Println("start finding slaves")
		foundedSlaves := core.Discovery.CheckNewSlaves()
		log.Println("founded services: ", foundedSlaves)
		core.SlaveMoniring.CompareAndSave(foundedSlaves)
		time.Sleep(time.Second * 15)
	}
}

/*UnregisterService - де регистрация сервиса из consul*/
func (core *MasterRunnerCore) UnregisterService() {
	core.Discovery.UnregisterCurrentService()
}

/*GetRouter - получение сконфигурированного роутера*/
func (core *MasterRunnerCore) GetRouter() *mux.Router {
	return core.RouterRunner.GetRouterMux()
}
