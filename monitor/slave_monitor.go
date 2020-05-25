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
		MaxExecutingTaskPerSlave int
	}

	/*Slave - configuration of slave available*/
	Slave struct {
		ID                  string
		Address             string
		Port                int
		CurrentExecuteTasks []int // index of SlaveMonitoring.CurrentTasks
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
func (slavemonitor *SlaveMonitoring) CompareAndSave(foundedServices []*consulapi.CatalogService) {
	for _, value := range foundedServices {
		if slavemonitor.notExistService(value) {
			log.Println("new service not exist in this master executor")
			slavemonitor.SlavesAvailable = append(slavemonitor.SlavesAvailable, Slave{
				ID:                  value.ServiceID,
				Address:             value.Address,
				Port:                value.ServicePort,
				CurrentExecuteTasks: []int{},
			})
		}
	}
}

func (slavemonitor *SlaveMonitoring) notExistService(service *consulapi.CatalogService) bool {
	for _, slave := range slavemonitor.SlavesAvailable {
		if slave.ID == service.ServiceID {
			log.Debug("service already exist by slave id: ", slave.ID)
			return false
		}
	}
	log.Info("service does not exist: ", service.ID)
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

// TaskResultFromSlave - обновление текущего статуса задачи со слейв модуля
func (slavemonitor *SlaveMonitoring) TaskResultFromSlave(payload payloads.ChangeStatusTask) error {
	for _, task := range slavemonitor.CurrentTasks {
		if task.ID == payload.TaskID {
			return slavemonitor.updateTaskStatus(payload.TaskID, models.TaskStatus{
				StatusIndex: models.TaskStatusIndx(payload.NewStatus),
				Stage:       payload.Stage,
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
		Status: models.TaskStatus{
			StatusIndex: models.QUEUED,
			Stage:       "",
		},
		SlaveIndex: slaveID,
	})
	slavemonitor.updateInfoInSlave(slaveID, len(slavemonitor.CurrentTasks)-1)
}

func (slavemonitor *SlaveMonitoring) updateInfoInSlave(slaveID int, taskID int) {
	slave := slavemonitor.SlavesAvailable[slaveID]
	slave.CurrentExecuteTasks = append(slave.CurrentExecuteTasks, taskID)
	slavemonitor.SlavesAvailable[slaveID] = slave
}

func (slavemonitor *SlaveMonitoring) updateTaskStatus(taskID string, newStatus models.TaskStatus) error {
	for index, task := range slavemonitor.CurrentTasks {
		timeFinish := time.Now().Unix()
		if newStatus.StatusIndex != models.FAILED && newStatus.StatusIndex != models.SUCCESS && newStatus.StatusIndex != models.CANCELED {
			timeFinish = -1
		}
		if result := slavemonitor.updateTasks(task.ID, taskID, index, newStatus, timeFinish); result {
			return nil
		}
	}
	return errors.New("can not update work status by undefined task")
}

// return true if task was updated, else return false
func (slavemonitor *SlaveMonitoring) updateTasks(taskIDCycle string, taskIDUpdatable string, currentTaskIDx int, status models.TaskStatus, timeFinished int64) bool {
	if taskIDCycle == taskIDUpdatable {
		currentTask := slavemonitor.CurrentTasks[currentTaskIDx]
		slavemonitor.CurrentTasks[currentTaskIDx] = models.Task{
			ID:            taskIDCycle,
			TimeCreated:   currentTask.TimeCreated,
			TimeFinishing: timeFinished,
			SlaveIndex:    currentTask.SlaveIndex,
			Status:        status,
		}
		return true
	}
	return false
}

/*GetTaskStatus - получить текущий статус задачи по её идентификатору*/
func (slavemonitor *SlaveMonitoring) GetTaskStatus(taskID string) (*models.TaskStatus, error) {
	for _, task := range slavemonitor.CurrentTasks {
		if task.ID == taskID {
			return &task.Status, nil
		}
	}
	return nil, errors.New("can not get task status by undefined task")
}
