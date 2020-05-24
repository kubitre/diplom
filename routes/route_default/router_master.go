package route_default

import (
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/enhancer"
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/monitor"
	"github.com/kubitre/diplom/payloads"
	"github.com/kubitre/diplom/routes"
	"github.com/kubitre/diplom/services"
	log "github.com/sirupsen/logrus"
)

//MasterRunnerRouter - main router for master runner
type MasterRunnerRouter struct {
	Router  *mux.Router
	service *services.MasterRunnerService
}

// InitializeMasterRunnerRouter - инициализация роутера мастер ноды
func InitializeMasterRunnerRouter(slaveMonitor *monitor.SlaveMonitoring) *MasterRunnerRouter {
	return &MasterRunnerRouter{
		Router: mux.NewRouter(),
	}
}

func initialRoutesSetup(router *mux.Router) *mux.Router {
	return router
}

// CreateNewTask - создание новой задачи на обработку репозитория кандидата post {workID, work by spec}
func (route *MasterRunnerRouter) CreateNewTask(writer http.ResponseWriter, request *http.Request) {
	var createNewTaskPayload payloads.CreateNewTask
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(&createNewTaskPayload); err != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "createNewTask",
			},
			"detailed": map[string]string{
				"message": "can't unmarshal into new work model",
				"trace":   err.Error(),
			},
		}, http.StatusBadRequest)
		return
	}

	if errRedirect := route.SlaveMonitoring.SendSlaveTask(request, writer, createNewTaskPayload); errRedirect != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "slaves",
				"func":    "SendSlaveTask",
			},
			"detailed": map[string]string{
				"message": "can't redirect new task into slave executor",
				"trace":   errRedirect.Error(),
			},
		}, http.StatusInternalServerError)
		return
	}
	enhancer.Response(request, writer, map[string]interface{}{
		"context": map[string]string{
			"module":  "master_executor",
			"package": "routers",
			"func":    "createNewTask",
		},
		"detailed": map[string]string{
			"message": "something error",
		},
	}, http.StatusInternalServerError)
	return
}

//ChangeTaskStatus - изменить текущий статус работы (остановить, запустить) post {taskID, status: [STARTED, STOPING, FINISHING, FAILED]}
func (route *MasterRunnerRouter) ChangeTaskStatus(writer http.ResponseWriter, request *http.Request) {
	var statusTaskChangePayload payloads.ChangeStatusTask
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(&statusTaskChangePayload); err != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "changeTaskStatus",
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
				"func":    "changeTaskStatus",
			},
			"detailed": map[string]string{
				"message": "can not be update status of task by unknown status",
				"trace":   errValide.Error(),
			},
		}, http.StatusBadRequest)
		return
	}
	if errUpdating := route.SlaveMonitoring.TaskResultFromSlave(statusTaskChangePayload); errUpdating != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "changeTaskStatus",
			},
			"detailed": map[string]string{
				"message": "can not be update status",
				"trace":   errUpdating.Error(),
			},
		}, http.StatusInternalServerError)
		return
	}
	enhancer.Response(request, writer, map[string]interface{}{
		"status": "update status was completed",
	}, http.StatusOK)
	return
}

// GetLogTask - получение логов с работы get ?taskID=:taskID&stage?=:nameStage
func (route *MasterRunnerRouter) GetLogTask(writer http.ResponseWriter, request *http.Request) {
	// get result task by his id
	log.Println("start working with getting log")
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	stage := vars["stage"]
	job := vars["job"]
	log.Println("taskID: "+taskID+"stage name: "+stage, " job: ", job)
	resultFile, errPreparing := enhancer.Mergelog(route.Config.PathToLogsWork, taskID, stage, job)
	if errPreparing != nil {
		log.Println("can not preparing log: ", errPreparing)
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "getLogTask",
			},
			"detailed": map[string]string{
				"message": "can not be merged logs",
				"trace":   errPreparing.Error(),
			},
		}, http.StatusBadRequest)
		return
	}
	writer.Header().Set("Content-Type", mime.TypeByExtension(resultFile))
	log.Println("start serving log: " + resultFile)
	http.ServeFile(writer, request, resultFile)
	return
}

func (route *MasterRunnerRouter) getAllLogsTree(writer http.ResponseWriter, request *http.Request) {
	log.Println("start working with getting all logs")
	writer.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext("logs/workid/stage1.log")))
	http.ServeFile(writer, request, "logs/workid/stage1.log")
	// enhancer.Response(request, writer, map[string]interface{}{
	// 	"status": "not implemented yet",
	// }, http.StatusNotImplemented)
	return
}

// CreateLogTask - создание логов с выполненной работы post {taskID, stage, logcontent}
func (route *MasterRunnerRouter) CreateLogTask(writer http.ResponseWriter, request *http.Request) {
	log.Println("start creating new log")
	var model models.LogsPerTask
	if err := json.NewDecoder(request.Body).Decode(&model); err != nil {
		log.Println("can not parsed body: ", err)
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "createLogWork",
			},
			"detailed": map[string]string{
				"message": "can't parse body",
				"trace":   err.Error(),
			},
		}, http.StatusBadRequest)
		return
	}
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	stage := vars["stage"]
	job := vars["job"]
	logPath := route.Config.PathToLogsWork + "/" + taskID + "/" + stage
	log.Println("create log path: ", logPath)
	errDirCreating := os.MkdirAll(logPath, os.ModePerm)
	if errDirCreating != nil {
		log.Println("can not be creating dir for log: ", errDirCreating)
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
	_, err := os.Create(logPath + "/" + job + ".log")
	if err != nil {
		log.Println("can not create log file: ", err)
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
	marsh, _ := json.Marshal(&model)

	// write data to job, append data to stage
	ioutil.WriteFile(logPath+"/"+job+".log", marsh, 0644)

	enhancer.Response(request, writer, map[string]interface{}{
		"status": "completed create log task by taskID and stage name",
	}, http.StatusOK)
	return
}

// GetTaskStatus - получение статуса задачи GET /taskID=:taskID
func (route *MasterRunnerRouter) GetTaskStatus(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	if taskID == "" {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "getTaskStatus",
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
				"func":    "getTaskStatus",
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

// GetStatusWorkers -  получение текущего статуса всех slave нод
func (route *MasterRunnerRouter) GetStatusWorkers(writer http.ResponseWriter, request *http.Request) {
	enhancer.Response(request, writer, map[string]interface{}{
		"available": route.SlaveMonitoring.SlavesAvailable,
	}, http.StatusOK)
}

// GetReportsPerTask - получение отчётов по задаче
func (route *MasterRunnerRouter) GetReportsPerTask(writer http.ResponseWriter, request *http.Request) {

}

// healthcheck - статус сервиса для service discovery
func (route *MasterRunnerRouter) healthCheck(writer http.ResponseWriter, request *http.Request) {
	// implement logic for return current running works and amount slaves
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("status is running"))
}

// обработка неизвестных запросов
func (route *MasterRunnerRouter) notFoundHandler(writer http.ResponseWriter, request *http.Request) {
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
func (route *MasterRunnerRouter) ConfiguringRoutes() {
	route.Router.HandleFunc(routes.ApiTaskCreate, route.CreateNewTask).Methods(http.MethodPost)
	route.Router.HandleFunc(routes.ApiTaskChangeOrGetStatus, route.ChangeTaskStatus).Methods(http.MethodPost)
	route.Router.HandleFunc(routes.ApiTaskChangeOrGetStatus, route.GetTaskStatus).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskLogJob, route.CreateLogTask).Methods(http.MethodPost)
	route.Router.HandleFunc(routes.ApiTaskLogJob, route.GetLogTask).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskLogStage, route.GetLogTask).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskLogTask, route.GetLogTask).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskLogAll, route.getAllLogsTree).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiAvailableWorkers, route.GetStatusWorkers).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskReport, route.GetReportsPerTask).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiHealthCheck, route.healthCheck).Methods(http.MethodGet)
	route.Router.NotFoundHandler = http.HandlerFunc(route.notFoundHandler)
	log.Println("Current configuration: ", route.Config)
}

/*GetRouterMux - получить сконфигурированный роутер*/
func (route *MasterRunnerRouter) GetRouterMux() *mux.Router {
	return route.Router
}
