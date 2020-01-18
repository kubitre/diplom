package routes

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/runner"
)

// main router for master runner
type RunnerRouter struct {
	Router *mux.Router
	Runner *runner.MasterNodeRunner
}

const (
	apiInitializeConfigure  = "/setting_runner"
	apiTask                 = "/tasks"
	apiGetCurrentTaskStatus = "/status/tasks"
)

// create runner router
func InitializeRunnerRouter() *RunnerRouter {
	return &RunnerRouter{
		Router: mux.NewRouter(),
		Runner: runner.InitializeMasterNode(),
	}
}

func initialRoutesSetup(router *mux.Router) *mux.Router {
	return router
}

func (route *RunnerRouter) createNewTask(writer *http.ResponseWriter, request *http.Request) {
	// creating new task
}

func (route *RunnerRouter) stopTaskByID(writer *http.ResponseWriter, request *http.Request) {
	// stoping work task on runner
}

func (route *RunnerRouter) getTaskReportByTaskID(writer *http.ResponseWriter, request *http.Request) {
	// get result task by his id
}

func (route *RunnerRouter) getCurrentTasks(writer *http.ResponseWriter, request *http.Request) {
	// get current running tasks
}

func (route *RunnerRouter) initializeRunner(writer *http.ResponseWriter, request *http.Request) {
	// initialize runner with setting payload
}
