package enhancer

import (
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/monitor"
	"github.com/kubitre/diplom/payloads"
)

func MergeTasksWithSlaves(slaves []monitor.Slave, tasks []models.Task, executingTask, history []int) []payloads.EnhancedSlave {
	result := []payloads.EnhancedSlave{}
	for _, slave := range slaves {
		result = append(result, mergeTasksWithSlave(slave, tasks, executingTask, history))
	}
	return result
}

func mergeTasksWithSlave(slave monitor.Slave, tasks []models.Task, executedTask, historyTasks []int) payloads.EnhancedSlave {
	result := []models.Task{}
	history := []models.Task{}
	for _, taskID := range slave.CurrentExecuteTasks {
		for _, taskIndex := range executedTask {
			if taskIndex == taskID {
				result = append(result, tasks[taskIndex])
			}
		}
	}
	for _, taskID := range slave.HistoryTasks {
		for _, taskIndex := range historyTasks {
			if taskIndex == taskID {
				history = append(result, tasks[taskIndex])
			}
		}
	}

	return payloads.EnhancedSlave{
		ID:                  slave.ID,
		Address:             slave.Address,
		Port:                slave.Port,
		CurrentExecuteTasks: models.ConvertArrayTasks(result),
		HistoryExecuted:     models.ConvertArrayTasks(history),
	}
}
