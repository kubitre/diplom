package route_portal

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/enhancer"
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/payloads"
	"github.com/kubitre/diplom/routes"
	"github.com/kubitre/diplom/services"
)

//MasterRunnerRouterPortal - main router for master runner for portal adaptation
type MasterRunnerRouterPortal struct {
	Router  *mux.Router
	service *services.MasterRunnerService
}

// InitializeMasterRunnerRouter - инициализация роутера мастер ноды
func InitializeMasterRunnerRouter(masterService *services.MasterRunnerService) *MasterRunnerRouterPortal {
	return &MasterRunnerRouterPortal{
		Router:  mux.NewRouter(),
		service: masterService,
	}
}

func initialRoutesSetup(router *mux.Router) *mux.Router {
	return router
}

// CreateNewTask - создание новой задачи на обработку репозитория кандидата post {workID, work by spec}
func (route *MasterRunnerRouterPortal) CreateNewTask(writer http.ResponseWriter, request *http.Request) {
	var createNewTaskPayload models.TaskConfig
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
	route.service.NewTask(&createNewTaskPayload, request, writer)

}

//ChangeTaskStatus - изменить текущий статус работы (остановить, запустить) post {taskID, status: [STARTED, STOPING, FINISHING, FAILED]}
func (route *MasterRunnerRouterPortal) ChangeTaskStatus(writer http.ResponseWriter, request *http.Request) {
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
	route.service.ChangeStatusTask(&statusTaskChangePayload, request, writer)
}

// GetLogTask - получение логов с работы get ?taskID=:taskID&stage?=:nameStage
func (route *MasterRunnerRouterPortal) GetLogTask(writer http.ResponseWriter, request *http.Request) {
	log.Println("start working with getting log")
	route.service.GetLogsPerTask(request, writer)
}

// на стабилизацию
func (route *MasterRunnerRouterPortal) getAllLogsTree(writer http.ResponseWriter, request *http.Request) {
	log.Println("start working with getting all logs")
	writer.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext("logs/workid/stage1.log")))
	http.ServeFile(writer, request, "logs/workid/stage1.log")
	// enhancer.Response(request, writer, map[string]interface{}{
	// 	"status": "not implemented yet",
	// }, http.StatusNotImplemented)
	return
}

// CreateLogTask - создание логов с выполненной работы post {taskID, stage, logcontent}
func (route *MasterRunnerRouterPortal) CreateLogTask(writer http.ResponseWriter, request *http.Request) {
	log.Println("start creating new log")
	route.service.CreateLogTask(request, writer)
}

// GetTaskStatus - получение статуса задачи GET /taskID=:taskID
func (route *MasterRunnerRouterPortal) GetTaskStatus(writer http.ResponseWriter, request *http.Request) {
	route.service.GetTaskStatus(request, writer)
}

// GetStatusWorkers -  получение текущего статуса всех slave нод
func (route *MasterRunnerRouterPortal) GetStatusWorkers(writer http.ResponseWriter, request *http.Request) {
	route.service.GetStatusWorkers(request, writer)
}

// GetReportsPerTask - получение отчётов по задаче
func (route *MasterRunnerRouterPortal) GetReportsPerTask(writer http.ResponseWriter, request *http.Request) {
	route.service.GetReportPerTask(request, writer)
}

// healthcheck - статус сервиса для service discovery
func (route *MasterRunnerRouterPortal) healthCheck(writer http.ResponseWriter, request *http.Request) {
	// implement logic for return current running works and amount slaves
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("status is running"))
}

// обработка неизвестных запросов
func (route *MasterRunnerRouterPortal) notFoundHandler(writer http.ResponseWriter, request *http.Request) {
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

/*ConfigureRouter - конфигурирование маршрутов
 */
func (route *MasterRunnerRouterPortal) ConfigureRouter() {
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
}

/*GetRouter - получить сконфигурированный роутер*/
func (route *MasterRunnerRouterPortal) GetRouter() *mux.Router {
	return route.Router
}