package models

type (
	/*Task - description for task*/
	Task struct {
		ID            string
		SlaveIndex    int
		Status        TaskStatus
		TimeCreated   int64
		TimeFinishing int64
	}

	/*TaskStatus - state of task*/
	TaskStatus struct {
		StatusIndex TaskStatusIndx
		Stage       string
	}

	// TaskStatusIndx - индекс текущого статуса
	TaskStatusIndx int
)

const (
	// QUEUED - task insert in master executor and sending to
	QUEUED TaskStatusIndx = 1
	// RUNNING - task start in slave executor
	RUNNING = 2
	// CANCELED - task was stopped by client (like default plugin or portal)
	CANCELED = 3
	// FAILED - task was failed
	FAILED = 4 // task was failed
	// SUCCESS - task was successfully
	SUCCESS = 5 // task was successfull
)
