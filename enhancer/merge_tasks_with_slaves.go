package enhancer

import (
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/monitor"
	"github.com/kubitre/diplom/payloads"
)

func MergeTasksWithSlaves(slaves []monitor.Slave, tasks []models.Task, history []models.Task) []payloads.EnhancedSlave {
	result := []payloads.EnhancedSlave{}
	for _, slave := range slaves {
		result = append(result, mergeTasksWithSlave(slave, tasks, history))
	}
	return result
}

func mergeTasksWithSlave(slave monitor.Slave, tasks []models.Task, historyTasks []models.Task) payloads.EnhancedSlave {
	result := []models.Task{}
	history := []models.Task{}
	for _, taskID := range slave.CurrentExecuteTasks {
		for taskid, task := range tasks {
			if taskid == taskID {
				result = append(result, task)
			}
		}
	}
	for _, taskID := range slave.HistoryTasks {
		for taskid, task := range historyTasks {
			if taskid == taskID {
				history = append(result, task)
			}
		}
	}
	return payloads.EnhancedSlave{
		ID:                  slave.ID,
		Address:             slave.Address,
		Port:                slave.Port,
		CurrentExecuteTasks: result,
		HistoryExecuted:     history,
	}
}
