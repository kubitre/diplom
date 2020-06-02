package monitor

import (
	"bytes"
	"errors"
	"net/http"
	"strconv"
	"time"

	consulapi "github.com/hashicorp/consul/api"
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/payloads"
	log "github.com/sirupsen/logrus"
)

type (
	/*SlaveMonitoring - monitoring for current available workers, current state of tasks*/
	SlaveMonitoring struct {
		SlavesAvailable          []Slave
		LastUsingService         int
		CurrentTasks             []models.Task
		History                  []models.Task // TODO: Change from currentTasks to this place
		MaxExecutingTaskPerSlave int
	}

	/*Slave - configuration of slave available*/
	Slave struct {
		ID                  string
		Address             string
		Port                int
		CurrentExecuteTasks []int // index of SlaveMonitoring.CurrentTasks
		HistoryTasks        []int // index of executed tasks
	}

	// SlaveStatus - статус слейв модуля
	SlaveStatus int
)

const (
	// NOTEXISTSTAGE - не существующая стадия
	NOTEXISTSTAGE = "NOT_A_STAGE_()()"
)

/*InitializeNewSlaveMonitoring - инициализация части мониторинга слейв модулей*/
func InitializeNewSlaveMonitoring(maxTaskPerSlave int) (*SlaveMonitoring, error) {
	if maxTaskPerSlave == 0 {
		return nil, errors.New("minimal task executing per slave is 1")
	}
	return &SlaveMonitoring{
		MaxExecutingTaskPerSlave: maxTaskPerSlave,
		LastUsingService:         0,
	}, nil
}

/*CompareAndSave - сравнение с текущими узлами слейв и вновь полученными*/
func (slavemonitor *SlaveMonitoring) CompareAndSave(foundedServices []*consulapi.ServiceEntry) {
	for _, value := range foundedServices {
		if slavemonitor.notExistService(value) {
			log.Println("new service not exist in this master executor")
			slavemonitor.SlavesAvailable = append(slavemonitor.SlavesAvailable, Slave{
				ID:                  value.Service.ID,
				Address:             value.Node.Address,
				Port:                value.Service.Port,
				CurrentExecuteTasks: []int{},
			})
		}
	}
}

/*ClearNotAvailableSlaves - отчистка недоступных слейвов*/
func (slavemonitor *SlaveMonitoring) ClearNotAvailableSlaves(available []*consulapi.ServiceEntry) {
	result := []Slave{}
	for _, monitoredServices := range slavemonitor.SlavesAvailable {
		for _, availableService := range available {
			if monitoredServices.ID == availableService.Service.ID {
				result = append(result, monitoredServices)
			}
		}
	}
	slavemonitor.SlavesAvailable = result
}

func (slavemonitor *SlaveMonitoring) notExistService(service *consulapi.ServiceEntry) bool {
	for _, slave := range slavemonitor.SlavesAvailable {
		if slave.ID == service.Service.ID {
			log.Debug("service already exist by slave id: ", slave.ID)
			return false
		}
	}
	log.Info("service does not exist: ", service.Service.ID)
	return true
}

// SendSlaveTask - проксирование запроса от клиента на один из слейв сервисов
func (slavemonitor *SlaveMonitoring) SendSlaveTask(request *http.Request, writer http.ResponseWriter, newTask *models.TaskConfig) error {
	if newTask.TaskID == "" {
		return errors.New("value of taskID can not be null or empty")
	}
	log.Debug("start chosing slave executor")
	// need refactoring
	slaveID, err := slavemonitor.chooseHaveSpaceForWorkSlave()
	if err != nil {
		return err
	}
	body, err := newTask.ToByteArray()
	if err != nil {
		return err
	}
	rbody := bytes.NewReader(body)
	log.Debug("choosed slave: ", slaveID)
	slavemonitor.addNewTask(newTask.TaskID, slaveID)
	addressSlave := "http://" + slavemonitor.SlavesAvailable[slaveID].Address + ":" + strconv.Itoa(slavemonitor.SlavesAvailable[slaveID].Port)
	log.Debug("starting redirect to : ", addressSlave)
	_, err = http.Post(addressSlave+"/task", "application/json", rbody)
	if err != nil {
		return err
	}

	return nil
}

// func (slavemonitor *SlaveMonitoring) garbageTaskCollector() {
// 	for _, task := range slavemonitor.CurrentTasks {
// 		if  task.TimeCreated
// 	}
// }

// func (slavemonitor *SlaveMonitoring) checkTaskComplete(taskID int) {
// 	time.Sleep(time.Minute * 1)
// 	statusIndex := slavemonitor.CurrentTasks[taskID].Status.StatusIndex
// 	if slavemonitor.CurrentTasks[taskID].Status.StatusIndex != models.SUCCESS ||
// }

// TaskResultFromSlave - обновление текущего статуса задачи со слейв модуля
func (slavemonitor *SlaveMonitoring) TaskResultFromSlave(payload payloads.ChangeStatusTask) error {
	for _, task := range slavemonitor.CurrentTasks {
		if task.ID == payload.TaskID {
			return slavemonitor.updateTaskStatus(payload.TaskID, models.TaskStatusIndx(payload.NewStatus), payload.CurrentStage)
		}
	}

	return errors.New("can not find task by taskID: " + payload.TaskID)
}

// JobResultFromSlave - обновление текущего статуса job со слейв модуля
func (slavemonitor *SlaveMonitoring) JobResultFromSlave(payload *payloads.ChangeStatusJob) error {
	log.Info("started update job status")
	for _, task := range slavemonitor.CurrentTasks {
		if task.ID == payload.TaskID {
			log.Info("start updated job status")
			return slavemonitor.updateJobStatus(task.ID, models.JobStatus{
				Job:         payload.Job,
				StatusIndex: models.TaskStatusIndx(payload.NewStatus),
			})
		}
	}
	return errors.New("can not find task by taskID: " + payload.TaskID)
}

func (slavemonitor *SlaveMonitoring) chooseHaveSpaceForWorkSlave() (int, error) {
	if len(slavemonitor.SlavesAvailable) == 0 {
		return -1, errors.New("can not execute this task, because not have any available slave executors")
	}
	currentUseSlaveIndex := 0
	lastUsedIndex := slavemonitor.LastUsingService
	if lastUsedIndex == len(slavemonitor.SlavesAvailable)-1 {
		slavemonitor.changeLastIndex(currentUseSlaveIndex)
		return currentUseSlaveIndex, nil
	}
	slavemonitor.changeLastIndex(lastUsedIndex + 1)
	return lastUsedIndex + 1, nil
}

func (slavemonitor *SlaveMonitoring) changeLastIndex(newIndex int) {
	slavemonitor.LastUsingService = newIndex
}

func (slavemonitor *SlaveMonitoring) updateOneOfSlave(index int, currentExecutingTask int) {
	slave := slavemonitor.SlavesAvailable[index]
	slavemonitor.SlavesAvailable[index] = Slave{
		ID:                  slave.ID,
		Address:             slave.Address,
		CurrentExecuteTasks: append(slave.CurrentExecuteTasks, currentExecutingTask),
		Port:                slave.Port,
	}
}

func (slavemonitor *SlaveMonitoring) addNewTask(taskID string, slaveID int) {
	slavemonitor.CurrentTasks = append(slavemonitor.CurrentTasks, models.Task{
		ID:          taskID,
		TimeCreated: time.Now().Unix(),
		StatusJobs:  []models.JobStatus{},
		StatusTask:  models.QUEUED,
		SlaveIndex:  slaveID,
	})
	slavemonitor.updateInfoInSlave(slaveID, len(slavemonitor.CurrentTasks)-1)
}

func (slavemonitor *SlaveMonitoring) updateInfoInSlave(slaveID int, taskID int) {
	slave := slavemonitor.SlavesAvailable[slaveID]
	slave.CurrentExecuteTasks = append(slave.CurrentExecuteTasks, taskID)
	slavemonitor.SlavesAvailable[slaveID] = slave
}

func (slavemonitor *SlaveMonitoring) updateTaskStatus(taskID string, newStatus models.TaskStatusIndx, stage string) error {
	for index, task := range slavemonitor.CurrentTasks {
		timeFinish := time.Now().Unix()
		if newStatus != models.FAILED && newStatus != models.SUCCESS && newStatus != models.CANCELED {
			timeFinish = -1
		}
		if result := slavemonitor.updateTasks(task.ID, taskID, index, newStatus, timeFinish, stage); result {
			if newStatus != models.RUNNING && newStatus != models.QUEUED {
				log.Debug("starting update current task to history")
				slavemonitor.updateHistory(taskID)
				slavemonitor.updateExecutedTaskPerSlaves(index)
			}
			return nil
		}
	}
	return errors.New("can not update task status by undefined task")
}

/*CheckTaskIDExist - проверка, что задача с таким идентификатором существует уже*/
func (slavemonitor *SlaveMonitoring) CheckTaskIDExist(taskID string) bool {
	for _, task := range append(append([]models.Task{}, slavemonitor.CurrentTasks...), slavemonitor.History...) {
		if task.ID == taskID {
			return true
		}
	}
	return false
}

func (slavemonitor *SlaveMonitoring) updateJobStatus(taskID string, newStatus models.JobStatus) error {
	for index, task := range slavemonitor.CurrentTasks {
		timeFinish := time.Now().Unix()
		log.Debug("Time Finish job with status: ", newStatus.StatusIndex.GetString(), " time: ", timeFinish)
		if newStatus.StatusIndex != models.FAILED && newStatus.StatusIndex != models.SUCCESS && newStatus.StatusIndex != models.CANCELED {
			log.Debug("change time to -1")
			timeFinish = -1
		}
		if result := slavemonitor.updateJob(task.ID, taskID, index, newStatus, timeFinish); result {
			return nil
		}
	}
	return errors.New("can not update job status by undefined task")
}

func (slavemonitor *SlaveMonitoring) updateJob(taskID string, taskIDUpdatable string, currentTaskIDX int, jobStatus models.JobStatus, timeFinished int64) bool {
	if taskID == taskIDUpdatable {
		log.Info("updated job. Time: ", timeFinished)
		currentTask := slavemonitor.CurrentTasks[currentTaskIDX]
		log.Debug("task: ", currentTask)
		statusPerJobs := currentTask.StatusJobs
		log.Debug("jobs per task: ", statusPerJobs)
		updated := false
		for indexJob, job := range statusPerJobs {
			if job.Job == jobStatus.Job {
				statusPerJobs[indexJob] = models.JobStatus{
					Job:           job.Job,
					StatusIndex:   jobStatus.StatusIndex,
					TimeFinishing: timeFinished,
				}
				updated = true
			}
		}
		if !updated {
			jobStatus.TimeFinishing = timeFinished
			log.Info("append job to result ")
			statusPerJobs = append(statusPerJobs, jobStatus)
		}
		log.Info("jobs status: ", statusPerJobs)
		slavemonitor.CurrentTasks[currentTaskIDX] = models.Task{
			ID:            taskID,
			TimeCreated:   currentTask.TimeCreated,
			TimeFinishing: currentTask.TimeFinishing,
			SlaveIndex:    currentTask.SlaveIndex,
			StatusTask:    currentTask.StatusTask,
			StatusJobs:    statusPerJobs,
		}
		return true
	}
	return false
}

// return true if task was updated, else return false
func (slavemonitor *SlaveMonitoring) updateTasks(taskIDCycle string, taskIDUpdatable string, currentTaskIDx int, status models.TaskStatusIndx, timeFinished int64, stage string) bool {
	if taskIDCycle == taskIDUpdatable {
		currentTask := slavemonitor.CurrentTasks[currentTaskIDx]
		slavemonitor.CurrentTasks[currentTaskIDx] = models.Task{
			ID:            taskIDCycle,
			TimeCreated:   currentTask.TimeCreated,
			TimeFinishing: timeFinished,
			SlaveIndex:    currentTask.SlaveIndex,
			Stage:         stage,
			StatusTask:    status,
			StatusJobs:    currentTask.StatusJobs,
		}

		return true
	}
	return false
}

func (slavemonitor *SlaveMonitoring) updateHistory(taskID string) {
	result := []models.Task{}
	for _, value := range slavemonitor.CurrentTasks {
		if value.ID == taskID {
			slavemonitor.History = append(slavemonitor.History, value)
		} else {
			result = append(result, value)
		}
	}
	log.Debug("GLOBAL TASKS: ", result)
	log.Debug("GLOBAL HISTORY: ", slavemonitor.History)
	slavemonitor.CurrentTasks = result
}

func (slavemonitor *SlaveMonitoring) updateExecutedTaskPerSlaves(indexTask int) {
	for slaveID, slave := range slavemonitor.SlavesAvailable {
		currentTasks := []int{}
		currentHistroy := []int{}
		for _, slaveTask := range slave.CurrentExecuteTasks {
			if indexTask != slaveTask {
				currentTasks = append(currentTasks, slaveTask)
			} else {
				currentHistroy = append(currentHistroy, slaveTask)
			}
		}
		currentHistroy = append(currentHistroy, slavemonitor.SlavesAvailable[slaveID].HistoryTasks...)
		log.Debug("change histry and current tasks: History: ", currentHistroy, " Tasks:", currentTasks)
		slavemonitor.SlavesAvailable[slaveID].CurrentExecuteTasks = currentTasks
		slavemonitor.SlavesAvailable[slaveID].HistoryTasks = currentHistroy
		log.Debug("Updated task in slave: ", slavemonitor.SlavesAvailable[slaveID].HistoryTasks)
	}
}

/*GetTaskStatus - получить текущий статус задачи по её идентификатору*/
func (slavemonitor *SlaveMonitoring) GetTaskStatus(taskID string) (*models.Task, error) {
	for _, task := range slavemonitor.CurrentTasks {
		if task.ID == taskID {
			return &task, nil
		}
	}
	for _, task := range slavemonitor.History {
		if task.ID == taskID {
			return &task, nil
		}
	}
	return nil, errors.New("can not get task status by undefined task")
}
