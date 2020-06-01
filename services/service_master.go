package services

import (
	"bufio"
	"encoding/json"
	"io/ioutil"
	"mime"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/enhancer"
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/payloads"
	log "github.com/sirupsen/logrus"
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

/*GetReportPath - получить путь до текущих отчётов*/
func (service *MasterRunnerService) GetReportPath() string {
	return service.masterConfig.PathToReportsWork
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
}

// ChangeStatusJob - изменение статуса джобы
func (service *MasterRunnerService) ChangeStatusJob(statusTaskChangePayload *payloads.ChangeStatusJob, request *http.Request, writer http.ResponseWriter) {
	if errUpdating := service.masterCore.SlaveMoniring.JobResultFromSlave(statusTaskChangePayload); errUpdating != nil {
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
}

// GetLogsPerTask - получение логов по задаче (в случае если будет передан только taskID мержатся все логи из задачи, если будет taskID и stage - тогда только логи по стади и таске ну и по job в случае передачи taskID, stage, job)
func (service *MasterRunnerService) GetLogsPerTask(request *http.Request, writer http.ResponseWriter, taskID, stage, job string) {
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
	file, err := os.Create(logPath + "/" + job + ".log")
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
	service.textNotationLog(&model, file)
	enhancer.Response(request, writer, map[string]interface{}{
		"status": "completed create log task by taskID and stage name",
	}, http.StatusOK)
}

func (service *MasterRunnerService) textNotationLog(logs *models.LogsPerTask, file *os.File) {
	writer := bufio.NewWriter(file)
	if len(logs.STDOUT) > 0 {
		writer.WriteString("Stdout:\n")
		for _, line := range logs.STDOUT {
			writer.WriteString(line)
		}
		writer.WriteString("\n")
	}
	if len(logs.STDERR) > 0 {
		writer.WriteString("Stderr:\n")
		for _, line := range logs.STDERR {
			writer.WriteString(line)
		}
	}
	writer.Flush()
}

// GetTaskStatus - получить статус задачи по её идентификатору
func (service *MasterRunnerService) GetTaskStatus(request *http.Request, writer http.ResponseWriter, taskID string) *models.Task {
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
		return nil
	}
	return taskStatus
}

// GetStatusWorkers - получение текущего состояния всех воркеров
func (service *MasterRunnerService) GetStatusWorkers(request *http.Request, writer http.ResponseWriter) {
	enhancer.Response(request, writer, map[string]interface{}{
		"available": enhancer.MergeTasksWithSlaves(service.masterCore.SlaveMoniring.SlavesAvailable, service.masterCore.SlaveMoniring.CurrentTasks),
	}, http.StatusOK)
}

/*CreateReportsPerTask - запись отчётов по задаче*/
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
	log.Debug("Report: ", model)
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	job := vars["job"]
	reportPath := service.GetReportPath() + "/" + taskID
	errDirCreating := os.MkdirAll(reportPath, os.ModePerm)
	if errDirCreating != nil {
		log.Println("can not be creating dir for log: ", errDirCreating)
		enhancer.Response(request, writer, map[string]interface{}{
			"context": map[string]string{
				"module":  "master_executor",
				"package": "routers",
				"func":    "CreateReportsPerTask",
			},
			"detailed": map[string]string{
				"message": "can't create report path path",
				"trace":   errDirCreating.Error(),
			},
		}, http.StatusBadRequest)
		return
	}

	_, err := os.Create(reportPath + "/" + job + ".json")
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
	marshaling, errMarshal := json.Marshal(model)
	if errMarshal != nil {
		log.Error(errMarshal)
		return
	}
	log.Debug("Value after marshaling: ", string(marshaling))
	ioutil.WriteFile(reportPath+"/"+job+".json", marshaling, 0666)
	enhancer.Response(request, writer, map[string]interface{}{
		"status": "success writing reports",
	}, http.StatusNotImplemented)
}

// GetReportPerTask - получение отчёта по задаче (в случае, если в задаче использовались extra параметры, для выделения каких-либо метрик и т.д.)
func (service *MasterRunnerService) GetReportPerTask(request *http.Request, writer http.ResponseWriter) map[string][]string {
	vars := mux.Vars(request)
	taskID := vars["taskID"]
	job := vars["job"]
	result, errReportGetting := service.GetReportsForStatus(taskID, job)
	if errReportGetting != nil {
		enhancer.Response(request, writer, map[string]interface{}{
			"trace": errReportGetting.Error(),
		}, http.StatusConflict)
	}
	return result
}

func (service *MasterRunnerService) GetReportsTask(taskID string) (map[string][]string, error) {
	reportPath := service.GetReportPath() + "/" + taskID
	var files []string
	result := map[string][]string{}
	if err := filepath.Walk(reportPath, func(path string, info os.FileInfo, err error) error {
		if path == reportPath {
			return nil
		}
		files = append(files, path)
		return nil
	}); err != nil {
		return nil, err
	}
	for _, file := range files {
		log.Debug("started getting metrics from file: ", file)
		resJob, errMetrics := service.GetReportsForStatus(taskID, file)
		log.Debug("getting metrics: ", resJob)
		if errMetrics != nil {
			return nil, errMetrics
		}
		result = service.appendMap(result, resJob)
	}
	log.Debug("result metrics map: ", result)
	return result, nil
}

func (service *MasterRunnerService) appendMap(resultMap map[string][]string, additionalMap map[string][]string) map[string][]string {
	enhancedResult := map[string][]string{}
	for firstname, values := range resultMap {
		additionalValues := values
		for secondName, valuesSecond := range additionalMap {
			if secondName == firstname {
				additionalValues = append(additionalValues, valuesSecond...)
			}
		}
		enhancedResult[firstname] = additionalValues
	}
	for firstname, values := range additionalMap {
		additionalValues := values
		for secondName, valuesSecond := range enhancedResult {
			if secondName == firstname {
				additionalValues = append(additionalValues, valuesSecond...)
			}
		}
		enhancedResult[firstname] = additionalValues
	}
	return enhancedResult
}

func (service *MasterRunnerService) GetReportsForStatus(taskID, fileName string) (map[string][]string, error) {
	file, errOpen := os.Open(fileName)
	if errOpen != nil {
		log.Error("can not read report path: ", errOpen)
		return nil, errOpen
	}
	newReader := bufio.NewReader(file)
	readed, errReading := ioutil.ReadAll(newReader)
	if errReading != nil {
		log.Error("can not open file for reading: ", errReading)
		return nil, errReading
	}
	var model map[string][]string
	if errUnmarshal := json.Unmarshal(readed, &model); errUnmarshal != nil {
		log.Error("can not unmarshal readed file into model: ", errReading)
		return nil, errUnmarshal
	}
	return model, nil
}

/*GetAgentID - получение текущего идентификатора агента*/
func (service *MasterRunnerService) GetAgentID() string {
	return service.masterConfig.AgentID
}
