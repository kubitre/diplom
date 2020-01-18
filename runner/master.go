package runner

import (
	"github.com/kubitre/diplom/models"
)

type (
	// Master node for all runners
	MasterNodeRunner struct {
		Slaves             []*TaskSlave // all current slaves running and executing his tasks
		NextSlaveIDForWork int
	}
	TaskSlave struct {
		Task chan *models.Task
	}
)

// creating new master node
func InitializeMasterNode() *MasterNodeRunner {
	return &MasterNodeRunner{
		Slaves: nil,
	}
}

// chose slave for executing new tasks
func (master *MasterNodeRunner) TaskChallenger() error {
	// <- master.Slaves[master.NextSlaveIDForWork].Task
	return nil
}

func (master *MasterNodeRunner) Run(task chan *models.Task) error {
	return nil
}
