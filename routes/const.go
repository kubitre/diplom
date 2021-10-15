package routes

const (
	ApiConfig                = "/configuration"
	ApiWorkers               = "/workers"
	ApiAvailableWorkers      = ApiWorkers + "/status"
	ApiTask                  = "/task"
	ApiTaskCreate            = ApiTask
	ApiTaskChangeOrGetStatus = ApiTask + "/{taskID:\\w+}/status"
	ApiJobChangeOrGetStatus  = ApiTaskChangeOrGetStatus + "/{jobName:\\w+}"
	ApiTaskReport            = ApiTask + "/{taskID:\\w+}/reports/{job:\\w+}"
	ApiTaskLogJob            = ApiTask + "/{taskID:\\w+}/log/{stage:\\w+}/{job:\\w+}"
	ApiTaskLogStage          = ApiTask + "/{taskID:\\w+}/log/{stage:\\w+}"
	ApiTaskLogTask           = ApiTask + "/{taskID:\\w+}/log"

	ApiTaskLogAll = ApiTask + "/getlogs"

	ApiHealthCheck = "/health"

	ApiTasksView = ApiTask + "/all"
)
