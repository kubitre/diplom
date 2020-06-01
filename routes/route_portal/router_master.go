package route_portal

import (
	"encoding/json"
	"log"
	"mime"
	"net/http"
	"path/filepath"
	"time"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/enhancer"
	"github.com/kubitre/diplom/middlewares"
	"github.com/kubitre/diplom/payloads"
	"github.com/kubitre/diplom/portal_models"
	"github.com/kubitre/diplom/routes"
	"github.com/kubitre/diplom/services"
	"github.com/kubitre/diplom/tools"
	"github.com/kubitre/diplom/validators"
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

// CreateNewTask - создание новой задачи на обработку репозитория кандидата
func (route *MasterRunnerRouterPortal) CreateNewTask(writer http.ResponseWriter, request *http.Request) {
	var createNewTaskPayload portal_models.PortalTask
	defer request.Body.Close()
	if err := json.NewDecoder(request.Body).Decode(&createNewTaskPayload); err != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "createNewTask",
				"plugin":  "portal_hedgehog",
			},
			"detailed": map[string]string{
				"message": "can't unmarshal into new work model",
				"trace":   err.Error(),
			},
		}, http.StatusBadRequest)
		return
	}
	convertedTask := createNewTaskPayload.ConvertToAgentTask()
	route.service.NewTask(&convertedTask, request, writer)
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
				"plugin":  "portal_hedgehog",
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
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	stage := request.URL.Query().Get("job_group")
	job := request.URL.Query().Get("job")
	log.Println("taskID: "+taskID+"stage name: "+stage, " job: ", job)
	route.service.GetLogsPerTask(request, writer, taskID, stage, job)
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
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	if taskID == "" {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "services",
				"func":    "GetTaskStatus",
				"plugin":  "portal_hedgehog",
			},
			"detailed": map[string]string{
				"message": "taskID can not be empty or null",
			},
		}, http.StatusBadRequest)
		return
	}
	task := route.service.GetTaskStatus(request, writer, taskID)
	if task != nil {
		notEnhancedReports, errNotEnhanced := route.service.GetReportsTask(task.ID)
		runnerData := map[string]string{}
		resultData := map[string]string{}
		if task.TimeFinishing != -1 {
			resultData["elapsedtime"] = time.Unix(task.TimeFinishing, 0).Sub(time.Unix(task.TimeCreated, 0)).String()
		} else {
			resultData["elapsedtime"] = time.Now().Sub(time.Unix(task.TimeCreated, 0)).String()
		}
		if task.Stage == "unknown" {
			resultData["stage"] = "not working"
		} else {
			resultData["stage"] = task.Stage
		}

		if errNotEnhanced != nil {
			runnerData["reports"] = errNotEnhanced.Error()
		} else {
			reports, err := validators.ValidateMetricsForPortal(notEnhancedReports)
			if err != nil {
				runnerData["reports"] = err.Error()
			} else {
				runnerData = tools.AppendMap(runnerData, reports)
			}
			runnerData = tools.AppendMap(runnerData, enhancer.MergeMetricsToString(notEnhancedReports))
		}

		statusEnhanced := portal_models.PortalTaskStatus{
			TaskID:             task.ID,
			TaskStatus:         task.StatusTask.GetString(),
			UserViewResultData: resultData,
			DeveloperOnlyData:  runnerData,
		}
		marshaled, errMarshaling := json.Marshal(statusEnhanced)
		if errMarshaling != nil {
			enhancer.Response(request, writer, map[string]interface{}{
				"context": map[string]string{
					"module":  "master_executor",
					"package": "services",
					"func":    "GetTaskStatus",
					"plugin":  "portal_hedgehog",
				},
				"detailed": map[string]string{
					"message": "taskID can not be empty or null",
				},
			}, http.StatusBadRequest)
			return
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		writer.Write(marshaled)
	}
}

// GetStatusWorkers -  получение текущего статуса всех slave нод
func (route *MasterRunnerRouterPortal) GetStatusWorkers(writer http.ResponseWriter, request *http.Request) {
	route.service.GetStatusWorkers(request, writer)
}

// GetReportsPerTask - получение отчётов по задаче
func (route *MasterRunnerRouterPortal) GetReportsPerTask(writer http.ResponseWriter, request *http.Request) {
	metrics := route.service.GetReportPerTask(request, writer)
	enhancedMetrics, errorValidating := validators.ValidateMetricsForPortal(metrics)
	if errorValidating != nil {
		log.Println("can not validated metrics: ", enhancedMetrics)
		enhancer.Response(request, writer, map[string]interface{}{
			"status": "bad validating metrics",
			"trace":  errorValidating.Error(),
		}, http.StatusConflict)
		return
	}
	enhancer.Response(request, writer, map[string]interface{}{
		"data": enhancedMetrics,
	}, http.StatusOK)
}

// healthcheck - статус сервиса для service discovery
func (route *MasterRunnerRouterPortal) healthCheck(writer http.ResponseWriter, request *http.Request) {
	// implement logic for return current running works and amount slaves
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("status is running"))
}

func (route *MasterRunnerRouterPortal) agentVerification(writer http.ResponseWriter, request *http.Request) {
	enhancer.Response(request, writer, map[string]interface{}{
		"runner_id": route.service.GetAgentID(),
	}, http.StatusOK)
}

// обработка неизвестных запросов
func (route *MasterRunnerRouterPortal) notFoundHandler(writer http.ResponseWriter, request *http.Request) {
	enhancer.Response(request, writer, map[string]interface{}{
		"context": map[string]string{
			"module":  "master_executor",
			"package": "routers",
			"func":    "notFoundHandler",
			"plugin":  "portal_hedgehog",
		},
		"detailed": map[string]string{
			"message": "not founded handler for you request",
			"route":   request.URL.RequestURI(),
		},
	}, http.StatusNotFound)
}

// CreateReportsPerTask - создание метрик на задачу из слейва
func (route *MasterRunnerRouterPortal) CreateReportsPerTask(writer http.ResponseWriter, request *http.Request) {
	log.Println("start creating reports")
	route.service.CreateReportsPerTask(request, writer)
}

// ChangeJobStatus - изменить текущий статус конкретной джобы
func (route *MasterRunnerRouterPortal) ChangeJobStatus(writer http.ResponseWriter, request *http.Request) {
	log.Println("start change job status")
	var statusTaskChangePayload payloads.ChangeStatusJob
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
	route.service.ChangeStatusJob(&statusTaskChangePayload, request, writer)
}

/*ConfigureRouter - конфигурирование маршрутов
 */
func (route *MasterRunnerRouterPortal) ConfigureRouter() {
	route.Router.HandleFunc(APITask, middlewares.CheckAgentID(route.service.GetAgentID(), http.HandlerFunc(route.CreateNewTask))).Methods(http.MethodPost)
	route.Router.HandleFunc(routes.ApiTaskChangeOrGetStatus, route.ChangeTaskStatus).Methods(http.MethodPost)
	route.Router.HandleFunc(APIStatusTask, middlewares.CheckAgentID(route.service.GetAgentID(), http.HandlerFunc(route.GetTaskStatus))).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskLogJob, route.CreateLogTask).Methods(http.MethodPost)
	route.Router.HandleFunc(routes.ApiTaskLogJob, middlewares.CheckAgentID(route.service.GetAgentID(), http.HandlerFunc(route.GetLogTask))).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskLogStage, middlewares.CheckAgentID(route.service.GetAgentID(), http.HandlerFunc(route.GetLogTask))).Methods(http.MethodGet)
	route.Router.HandleFunc(ApILogsPerTask, middlewares.CheckAgentID(route.service.GetAgentID(), http.HandlerFunc(route.GetLogTask))).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskLogAll, middlewares.CheckAgentID(route.service.GetAgentID(), http.HandlerFunc(route.getAllLogsTree))).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiAvailableWorkers, middlewares.CheckAgentID(route.service.GetAgentID(), http.HandlerFunc(route.GetStatusWorkers))).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskReport, middlewares.CheckAgentID(route.service.GetAgentID(), http.HandlerFunc(route.GetReportsPerTask))).Methods(http.MethodGet)
	route.Router.HandleFunc(routes.ApiTaskReport, route.CreateReportsPerTask).Methods(http.MethodPost)
	route.Router.HandleFunc(routes.ApiJobChangeOrGetStatus, route.ChangeJobStatus).Methods(http.MethodPost)
	route.Router.HandleFunc(routes.ApiHealthCheck, route.healthCheck).Methods(http.MethodGet)
	route.Router.HandleFunc("/", route.agentVerification).Methods(http.MethodGet)
	route.Router.NotFoundHandler = http.HandlerFunc(route.notFoundHandler)
}

/*GetRouter - получить сконфигурированный роутер*/
func (route *MasterRunnerRouterPortal) GetRouter() *mux.Router {
	return route.Router
}
