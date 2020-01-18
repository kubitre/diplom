package runner

import "github.com/kubitre/diplom/models"

// Slave runner for executing all master task
type SlaveNodeRunner struct {
	CurrentTask *models.Task
}

// executing new request from master node runner
func (slave *SlaveNodeRunner) ExecuteNewTask(task *models.Task) error {
	return nil
}

func (slave *SlaveNodeRunner) Run(task *models.Task) error {
	return nil
}
