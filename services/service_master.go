package services

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"mime"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/enhancer"
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/payloads"
)

/*MasterRunnerService - сервис исполняющего модуля в режиме мастер*/
type MasterRunnerService struct {
	masterCore   *core.MasterRunnerCore
	masterConfig *config.ConfigurationMasterRunner
}

// GetCore - отдать текущее ядро
func (service *MasterRunnerService) GetCore() *core.MasterRunnerCore {
	return service.masterCore
}

/*InitializeMasterRunnerService - инициализация сервиса исполняющего модуля в режиме мастер*/
func InitializeMasterRunnerService(configService *config.ServiceConfig, masterConfig *config.ConfigurationMasterRunner) (*MasterRunnerService, error) {
	coreMaster, err := core.InitNewMasterRunnerCore(masterConfig, configService)
	if err != nil {
		return nil, err
	}
	coreMaster.Run()
	return &MasterRunnerService{
		masterCore:   coreMaster,
		masterConfig: masterConfig,
	}, nil
}

/*NewTask - создание задачи*/
func (service *MasterRunnerService) NewTask(taskConfig *models.TaskConfig, request *http.Request, writer http.ResponseWriter) {
	if errRedirect := service.masterCore.SlaveMoniring.SendSlaveTask(request, writer, taskConfig); errRedirect != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "monitor",
				"func":    "NewTask",
			},
			"detailed": map[string]string{
				"message": "can't redirect new task into slave executor",
				"trace":   errRedirect.Error(),
			},
		}, http.StatusInternalServerError)
		return
	}
	return
}

/*ChangeStatusTask - изменить статус задачи*/
func (service *MasterRunnerService) ChangeStatusTask(statusTaskChangePayload *payloads.ChangeStatusTask, request *http.Request, writer http.ResponseWriter) {
	if errValide := statusTaskChangePayload.Validate(); errValide != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "services",
				"func":    "ChangeStatusTask",
			},
			"detailed": map[string]string{
				"message": "can not be update status of task by unknown status",
				"trace":   errValide.Error(),
			},
		}, http.StatusBadRequest)
		return
	}
	if errUpdating := service.masterCore.SlaveMoniring.TaskResultFromSlave(*statusTaskChangePayload); errUpdating != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "monitor",
				"func":    "TaskResultFromSlave",
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

// GetLogsPerTask - получение логов по задаче (в случае если будет передан только taskID мержатся все логи из задачи, если будет taskID и stage - тогда только логи по стади и таске ну и по job в случае передачи taskID, stage, job)
func (service *MasterRunnerService) GetLogsPerTask(request *http.Request, writer http.ResponseWriter) {
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	stage := vars["stage"]
	job := vars["job"]
	log.Println("taskID: "+taskID+"stage name: "+stage, " job: ", job)
	resultFile, errPreparing := enhancer.Mergelog(service.masterConfig.PathToLogsWork, taskID, stage, job)
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

// CreateLogTask - закрытый метод разрешённый только для воркеров. Создание логов по задаче (по каждой конкретной job)
func (service *MasterRunnerService) CreateLogTask(request *http.Request, writer http.ResponseWriter) {
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
	logPath := service.masterConfig.PathToLogsWork + "/" + taskID + "/" + stage
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

// GetTaskStatus - получить статус задачи по её идентификатору
func (service *MasterRunnerService) GetTaskStatus(request *http.Request, writer http.ResponseWriter) {
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	if taskID == "" {
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "services",
				"func":    "GetTaskStatus",
			},
			"detailed": map[string]string{
				"message": "taskID can not be empty or null",
			},
		}, http.StatusBadRequest)
		return
	}
	taskStatus, err := service.masterCore.SlaveMoniring.GetTaskStatus(taskID)
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

// GetStatusWorkers - получение текущего состояния всех воркеров
func (service *MasterRunnerService) GetStatusWorkers(request *http.Request, writer http.ResponseWriter) {
	enhancer.Response(request, writer, map[string]interface{}{
		"available": enhancer.MergeTasksWithSlaves(service.masterCore.SlaveMoniring.SlavesAvailable, service.masterCore.SlaveMoniring.CurrentTasks),
	}, http.StatusOK)
}

func (service *MasterRunnerService) CreateReportsPerTask(request *http.Request, writer http.ResponseWriter) {
	var model map[string][]string
	if err := json.NewDecoder(request.Body).Decode(&model); err != nil {
		log.Println("can not parsed body: ", err)
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "services",
				"func":    "CreateReportsPerTask",
			},
			"detailed": map[string]string{
				"message": "can't parse body",
				"trace":   err.Error(),
			},
		}, http.StatusBadRequest)
		return
	}

	enhancer.Response(request, writer, map[string]interface{}{
		"status": "not implemented",
	}, http.StatusNotImplemented)
}

// GetReportPerTask - получение отчёта по задаче (в случае, если в задаче использовались extra параметры, для выделения каких-либо метрик и т.д.)
func (service *MasterRunnerService) GetReportPerTask(request *http.Request, writer http.ResponseWriter) {
	enhancer.Response(request, writer, map[string]interface{}{
		"status": "not implemented",
	}, http.StatusNotImplemented)
}
