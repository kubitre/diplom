package services

import (
	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/monitor"
	"github.com/kubitre/diplom/payloads"
)

type MasterRunnerService struct {
	masterCore      *core.MasterRunnerCore
	masterConfig    *config.ConfigurationMasterRunner
	slaveMonitoring *monitor.SlaveMonitoring
}

func InitializeMasterRunnerService(configService *config.ServiceConfig, masterConfig *config.ConfigurationMasterRunner) (*MasterRunnerService, error) {
	coreMaster, err := core.InitNewMasterRunnerCore(masterConfig, configService)
	if err != nil {
		return nil, err
	}
	return &MasterRunnerService{
		masterCore:   coreMaster,
		masterConfig: masterConfig,
	}, nil
}

func (service *MasterRunnerService) NewTask(taskConfig *models.TaskConfig) error {

}

func (service *MasterRunnerService) ChangeStatusTask(payload *payloads.ChangeStatusTask) error {

}

func (service *MasterRunnerService) GetLogsPerTask(taskID, stageName, jobName string) error {

}

func (service *MasterRunnerService) CreateLogTask(taskID, stageName, jobName string) error {

}

func (service *MasterRunnerService) GetTaskStatus(taskID string) error {

}

func (service *MasterRunnerService) GetStatusWorkers() error {

}
