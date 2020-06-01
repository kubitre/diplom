package models

type (
	/*Task - description for task*/
	Task struct {
		ID            string
		SlaveIndex    int
		StatusTask    TaskStatusIndx
		Stage         string
		StatusJobs    []JobStatus
		TimeCreated   int64
		TimeFinishing int64
	}

	/*JobStatus - статус выполненной\не выполненной джобы*/
	JobStatus struct {
		StatusIndex   TaskStatusIndx
		Job           string
		TimeFinishing int64
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

/*GetString - строковое представление статуса*/
func (taskStatus TaskStatusIndx) GetString() string {
	switch taskStatus {
	case QUEUED:
		return "queued"
	case RUNNING:
		return "started"
	case CANCELED:
		return "canceled"
	case FAILED:
		return "fail"
	case SUCCESS:
		return "success"
	default:
		return "unknown"
	}
}
