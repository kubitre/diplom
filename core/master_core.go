package core

import (
	"time"

	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/discovery"
	"github.com/kubitre/diplom/monitor"
	log "github.com/sirupsen/logrus"
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
	slaveMonitor.LastUsingService = 0
	return &MasterRunnerCore{
		SlaveMoniring: slaveMonitor,
		Discovery:     discovery.InitializeDiscovery(discovery.MasterPattern, configService),
	}, nil
}

/*Run - запуск роутера, discovery, получение информации о слейвах*/
func (core *MasterRunnerCore) Run() {
	core.Discovery.NewClientForConsule()
	core.Discovery.RegisterServiceWithConsul([]string{discovery.TagMaster})
	go core.checkerNewSlave()

}

func (core *MasterRunnerCore) checkerNewSlave() {
	for {
		log.Debug("start finding slaves")
		foundedSlaves := core.Discovery.GetService(discovery.SlavePattern, discovery.TagSlave)
		if foundedSlaves == nil {
			log.Debug("not founded health service slaves")
			time.Sleep(time.Second * 15)
			continue
		}
		log.Debug("founded services: ", foundedSlaves)
		core.SlaveMoniring.CompareAndSave(foundedSlaves)
		core.SlaveMoniring.ClearNotAvailableSlaves(foundedSlaves)
		time.Sleep(time.Second * 15)
	}
}

/*UnregisterService - де регистрация сервиса из consul*/
func (core *MasterRunnerCore) UnregisterService() {
	core.Discovery.UnregisterCurrentService()
}
