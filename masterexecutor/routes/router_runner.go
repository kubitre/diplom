package routes

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/masterexecutor/enhancer"
	"github.com/kubitre/diplom/masterexecutor/payloads"
	"github.com/kubitre/diplom/masterexecutor/slaves"
)

//RunnerRouter - main router for master runner
type RunnerRouter struct {
	Router          *mux.Router
	SlaveMonitoring *slaves.SlaveMonitoring
}

const (
	apiConfig                = "/configuration"
	apiWorkers               = "/workers"
	apiAvailableWorkers      = apiWorkers + "/status"
	apiTask                  = "/task"
	apiTaskCreate            = apiTask + "/create"
	apiTaskChangeOrGetStatus = apiTask + "/{taskID:\\w+}/status"
	apiTaskLog               = apiTask + "/{taskID:\\w+}/log/{stage:\\w+}"
	apiTaskLogAll            = apiTask + "/getlogs"

	apiHealthCheck = "/health"
)

// InitializeRunnerRouter - инициализация роутера мастер ноды
func InitializeRunnerRouter(slaveMonitor *slaves.SlaveMonitoring) *RunnerRouter {
	return &RunnerRouter{
		Router:          mux.NewRouter(),
		SlaveMonitoring: slaveMonitor,
	}
}

func initialRoutesSetup(router *mux.Router) *mux.Router {
	return router
}

// createNewWork - создание новой задачи на обработку репозитория кандидата post {workID, work by spec}
func (route *RunnerRouter) createNewWork(writer http.ResponseWriter, request *http.Request) {
	var createWorkPayload payloads.CreateNewWork
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(&createWorkPayload); err != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "createNewWork",
			},
			"detailed": map[string]string{
				"message": "can't unmarshal into new work model",
				"trace":   err.Error(),
			},
		}, http.StatusBadRequest)
		return
	}

	if errRedirect := route.SlaveMonitoring.SendSlaveWork(request, writer); errRedirect != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "createNewWork",
			},
			"detailed": map[string]string{
				"message": "can't redirect new task into slave executor",
				"trace":   errRedirect.Error(),
			},
		}, http.StatusBadRequest)
		return
	}

	return
}

// изменить текущий статус работы (остановить, запустить) post {taskID, status: [STARTED, STOPING, FINISHING, FAILED]}
func (route *RunnerRouter) changeWorkStatus(writer http.ResponseWriter, request *http.Request) {
	var statusTaskChangePayload payloads.ChangeStatusTask
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(&statusTaskChangePayload); err != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "changeWorkStatus",
			},
			"detailed": map[string]string{
				"message": "can't unmarshal into changeStatus model",
				"trace":   err.Error(),
			},
		}, http.StatusBadRequest)
		return
	}
	if errValide := statusTaskChangePayload.Validate(); errValide != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "changeWorkStatus",
			},
			"detailed": map[string]string{
				"message": "can not be update status of task by unknown status",
				"trace":   errValide.Error(),
			},
		}, http.StatusBadRequest)
		return
	}
	route.SlaveMonitoring.TaskResultFromSlave(statusTaskChangePayload)
	enhancer.Response(request, writer, map[string]interface{}{
		"status": "not implemented yet",
	}, http.StatusNotImplemented)
	return
}

// получение логов с работы get ?workid=:workID&stage?=:nameStage
func (route *RunnerRouter) getLogWork(writer http.ResponseWriter, request *http.Request) {
	// get result task by his id
	log.Println("start working with getting log")
	vars := mux.Vars(request)
	workID := vars["workid"]
	stage := vars["stage"]
	log.Println("workid: " + workID + "stage name: " + stage)
	files := workID
	if stage != "" {
		files += "_" + stage + ".log"
	}
	writer.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext("logs/"+files)))
	log.Println("start serving log: " + "logs/" + files)
	http.ServeFile(writer, request, "logs/"+files)
	return
}

func (route *RunnerRouter) getAllLogsTree(writer http.ResponseWriter, request *http.Request) {
	log.Println("start working with getting all logs")
	writer.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext("logs/workid/stage1.log")))
	http.ServeFile(writer, request, "logs/workid/stage1.log")
	// enhancer.Response(request, writer, map[string]interface{}{
	// 	"status": "not implemented yet",
	// }, http.StatusNotImplemented)
	return
}

// создание логов с выполненной работы post {workid, stage, logcontent}
func (route *RunnerRouter) createLogWork(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	workid := vars["workid"]
	stage := vars["stage"]
	errDirCreating := os.MkdirAll("logs/"+workid, os.ModePerm)
	if errDirCreating != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "createLogWork",
			},
			"detailed": map[string]string{
				"message": "can't create work id path",
				"trace":   errDirCreating.Error(),
			},
		}, http.StatusBadRequest)
		return
	}
	_, err := os.Create("logs/" + workid + "/" + stage + ".log")
	if err != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "createLogWork",
			},
			"detailed": map[string]string{
				"message": "can't create new log",
				"trace":   err.Error(),
			},
		}, http.StatusBadRequest)
		return
	}

	enhancer.Response(request, writer, map[string]interface{}{
		"status": "not implemented yet",
	}, http.StatusNotImplemented)
	return
}

// получение статуса работы над задачей get ?workid=:workID
func (route *RunnerRouter) getWorkStatus(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	if taskID == "" {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "getWorkStatus",
			},
			"detailed": map[string]string{
				"message": "taskID can not be empty or null",
			},
		}, http.StatusBadRequest)
		return
	}
	taskStatus, err := route.SlaveMonitoring.GetTaskStatus(taskID)
	if err != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "slaves",
				"func":    "GetTaskStatus",
			},
			"detailed": map[string]string{
				"message": "can not get task status",
				"trace":   err.Error(),
			},
		}, http.StatusInternalServerError)
		return
	}
	enhancer.Response(request, writer, map[string]interface{}{
		"status": taskStatus,
	}, http.StatusOK)
	return
}

// получение текущего статуса всех slave нод
func (route *RunnerRouter) getStatusWorkers(writer http.ResponseWriter, request *http.Request) {
	enhancer.Response(request, writer, map[string]interface{}{
		"available": route.SlaveMonitoring.SlavesAvailable,
	}, http.StatusOK)
}

// healthcheck - статус сервиса для service discovery
func (route *RunnerRouter) healthCheck(writer http.ResponseWriter, request *http.Request) {
	// implement logic for return current running works and amount slaves
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("status is running"))
}

// обработка неизвестных запросов
func (route *RunnerRouter) notFoundHandler(writer http.ResponseWriter, request *http.Request) {
	enhancer.Response(request, writer, map[string]interface{}{
		"context": map[string]string{
			"module":  "master_executor",
			"package": "routers",
			"func":    "notFoundHandler",
		},
		"detailed": map[string]string{
			"message": "not founded handler for you request",
		},
	}, http.StatusNotFound)
}

/*ConfiguringRoutes - конфигурирование маршрутов
 */
func (route *RunnerRouter) ConfiguringRoutes() {
	route.Router.HandleFunc(apiTaskCreate, route.createNewWork).Methods(http.MethodPost)
	route.Router.HandleFunc(apiTaskChangeOrGetStatus, route.changeWorkStatus).Methods(http.MethodPost)
	route.Router.HandleFunc(apiTaskChangeOrGetStatus, route.getWorkStatus).Methods(http.MethodGet)
	route.Router.HandleFunc(apiTaskLog, route.createLogWork).Methods(http.MethodPost)
	route.Router.HandleFunc(apiTaskLog, route.getLogWork).Methods(http.MethodGet)
	route.Router.HandleFunc(apiTaskLogAll, route.getAllLogsTree).Methods(http.MethodGet)
	route.Router.HandleFunc(apiAvailableWorkers, route.getStatusWorkers).Methods(http.MethodGet)

	route.Router.HandleFunc(apiHealthCheck, route.healthCheck).Methods(http.MethodGet)
	route.Router.NotFoundHandler = http.HandlerFunc(route.notFoundHandler)
}

/*GetRouterMux - получить сконфигурированный роутер*/
func (route *RunnerRouter) GetRouterMux() *mux.Router {
	return route.Router
}
