package core

import (
	"log"
	"time"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/discovery"
	"github.com/kubitre/diplom/routes"
)

/*MasterRunnerCore - ядро master ноды*/
type MasterRunnerCore struct {
	RouterRunner  *routes.MasterRunnerRouter
	Discovery     *discovery.Discovery
	SlaveMoniring *SlaveMonitoring
}

/*InitNewCore - инициализация ядра текущего сервиса*/
func InitNewCore(config *config.ConfigurationSlaveRunner,
	configService *config.ServiceConfig,
) (*MasterRunnerCore, error) {
	slaveMonitor, err := InitializeNewSlaveMonitoring(config.MaxTaskPerSlave)
	if err != nil {
		return nil, err
	}
	slaveMonitor.LastUsingService <- core.INIT_USED_SLAVE
	return &MasterRunnerCore{
		SlaveMoniring: slaveMonitor,
		RouterRunner:  routes.InitializeMasterRunnerRouter(slaveMonitor, config),
		Discovery:     discovery.InitializeDiscovery(discovery.MasterPattern, configService),
	}, nil
}

/*Run - запуск роутера, discovery, получение информации о слейвах*/
func (core *MasterRunnerCore) Run(config *config.ConfigurationMasterRunner) {
	core.Discovery.NewClientForConsule()
	core.Discovery.RegisterServiceWithConsul([]string{discovery.TagMaster})
	core.RouterRunner.ConfiguringRoutes()
	go core.checkerNewSlave()

}

func (core *MasterRunnerCore) checkerNewSlave() {
	for {
		log.Println("start finding slaves")
		// foundedSlaves := core.Discovery.CheckNewSlaves()
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
