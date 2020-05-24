package core

import (
	"log"
	"time"

	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/discovery"
	"github.com/kubitre/diplom/monitor"
)

/*MasterRunnerCore - ядро master ноды*/
type MasterRunnerCore struct {
	Discovery     *discovery.Discovery
	SlaveMoniring *monitor.SlaveMonitoring
}

/*InitNewMasterRunnerCore - инициализация ядра текущего сервиса*/
func InitNewMasterRunnerCore(config *config.ConfigurationMasterRunner,
	configService *config.ServiceConfig,
) (*MasterRunnerCore, error) {
	slaveMonitor, err := monitor.InitializeNewSlaveMonitoring(config.MaxTaskPerSlave)
	if err != nil {
		return nil, err
	}
	slaveMonitor.LastUsingService <- monitor.INIT_USED_SLAVE
	return &MasterRunnerCore{
		SlaveMoniring: slaveMonitor,
		Discovery:     discovery.InitializeDiscovery(discovery.MasterPattern, configService),
	}, nil
}

/*Run - запуск роутера, discovery, получение информации о слейвах*/
func (core *MasterRunnerCore) Run(config *config.ConfigurationMasterRunner) {
	core.Discovery.NewClientForConsule()
	core.Discovery.RegisterServiceWithConsul([]string{discovery.TagMaster})
	go core.checkerNewSlave()

}

func (core *MasterRunnerCore) checkerNewSlave() {
	for {
		log.Println("start finding slaves")
		foundedSlaves := core.Discovery.GetService(discovery.SlavePattern, discovery.TagSlave)
		log.Println("founded services: ", foundedSlaves)
		core.SlaveMoniring.CompareAndSave(foundedSlaves)
		time.Sleep(time.Second * 15)
	}
}

/*UnregisterService - де регистрация сервиса из consul*/
func (core *MasterRunnerCore) UnregisterService() {
	core.Discovery.UnregisterCurrentService()
}
