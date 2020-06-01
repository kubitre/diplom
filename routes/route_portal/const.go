package route_portal

const (
	APITask        = "/tasks"
	APIStatusTask  = APITask + "/{taskID:\\w+}"
	ApILogsPerTask = APITask + "/{taskID:\\w+}/log"
	ApiTaskReport  = APITask + "/{taskID:\\w+}/reports/{stage:\\w++}/{job:\\w+}"
)
