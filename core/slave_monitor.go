package core

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"strconv"
	"time"

	consulapi "github.com/hashicorp/consul/api"
)

type (
	/*SlaveMonitoring - monitoring for current available workers, current state of tasks*/
	SlaveMonitoring struct {
		SlavesAvailable          []Slave
		LastUsingService         chan int
		CurrentTasks             []Task
		MaxExecutingTaskPerSlave int
	}

	/*Task - description for task*/
	Task struct {
		ID            string
		SlaveIndex    int
		Status        TaskStatus
		TimeCreated   int64
		TimeFinishing int64
	}

	/*Slave - configuration of slave available*/
	Slave struct {
		ID                  string
		Address             string
		Port                int
		CurrentStatus       SlaveStatus
		CurrentExecuteTasks []int // index of SlaveMonitoring.CurrentTasks
	}

	/*TaskStatus - state of task*/
	TaskStatus struct {
		StatusIndex TaskStatusIndx
		Stage       string
	}

	TaskStatusIndx int
	SlaveStatus    int
)

const (
	NEW      TaskStatusIndx = 0 // task saved in portal
	QUEUED   TaskStatusIndx = 1 // task insert in master executor
	RUNNING                 = 2 // task start in slave executor
	CANCELED                = 3 // task was stopped by portal
	FAILED                  = 4 // task was failed
	SUCCESS                 = 5 // task was successfull
)

const (
	INIT_USED_SLAVE = -1
)

const (
	// WAITING_WORK - Slave сервис ожидает работы
	WAITING_WORK SlaveStatus = iota
	// SLAVE_WORKING - Slave сервис в данный момент выполняет задание
	SLAVE_WORKING = iota
)

/*InitializeNewSlaveMonitoring - инициализация части мониторинга слейв модулей*/
func InitializeNewSlaveMonitoring(maxTaskPerSlave int) (*SlaveMonitoring, error) {
	if maxTaskPerSlave == 0 {
		return nil, errors.New("minimal task executing per slave is 1")
	}
	return &SlaveMonitoring{
		LastUsingService:         make(chan int, 1),
		MaxExecutingTaskPerSlave: maxTaskPerSlave,
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
				CurrentStatus:       WAITING_WORK,
				CurrentExecuteTasks: []int{},
			})
			return
		}
	}
}

func (slavemonitor *SlaveMonitoring) notExistService(service *consulapi.CatalogService) bool {
	for _, slave := range slavemonitor.SlavesAvailable {
		if slave.ID == service.ServiceID {
			log.Println("service already exist by slave id: ", slave.ID)
			return false
		}
	}
	log.Println("service does not exist: ", service.ID)
	return true
}

func (slavemonitor *SlaveMonitoring) changeLastUsingServiceIndex(newindex int) {
	slavemonitor.LastUsingService <- newindex
}

// SendSlaveTask - проксирование запроса от клиента на один из слейв сервисов
func (slavemonitor *SlaveMonitoring) SendSlaveTask(request *http.Request, writer http.ResponseWriter, newTask payloads.CreateNewTask) error {
	if newTask.TaskID == "" {
		return errors.New("value of taskID can not be null or empty")
	}
	log.Println("start chosing slave executor")
	slaveID, err := slavemonitor.chooseHaveSpaceForWorkSlave()
	if err != nil {
		return err
	}
	body, err := newTask.ConvertToTaskConfigBytes()
	if err != nil {
		return err
	}
	rbody := bytes.NewReader(body)
	log.Println("choosed slave: ", slaveID)
	slavemonitor.addNewTask(newTask.TaskID, slaveID)
	addressSlave := "http://" + slavemonitor.SlavesAvailable[slaveID].Address + ":" + strconv.Itoa(slavemonitor.SlavesAvailable[slaveID].Port)
	log.Println("starting redirect to : ", addressSlave)
	http.Post(addressSlave+"/task", "application/json", rbody)
	return nil
}

// TaskResultFromSlave - обновление текущего статуса задачи со слейв модуля
func (slavemonitor *SlaveMonitoring) TaskResultFromSlave(payload payloads.ChangeStatusTask) error {
	for _, task := range slavemonitor.CurrentTasks {
		if task.ID == payload.TaskID {
			return slavemonitor.updateTaskStatus(payload.TaskID, TaskStatus{
				StatusIndex: TaskStatusIndx(payload.NewStatus),
				Stage:       payload.Stage,
			}, payload.TimeFinished)
		}
	}

	return errors.New("can not find task by taskID: " + payload.TaskID)
}

func (slavemonitor *SlaveMonitoring) chooseHaveSpaceForWorkSlave() (int, error) {
	currentUseSlaveIndex := 0
	lastUsedIndex := <-slavemonitor.LastUsingService
	if lastUsedIndex == -1 {
		if len(slavemonitor.SlavesAvailable) > 0 {
			slavemonitor.changeLastUsingServiceIndex(currentUseSlaveIndex)
			return currentUseSlaveIndex, nil
		}
		return INIT_USED_SLAVE, errors.New("not installed slaves executors")
	}
	return slavemonitor.roundRobin()
}

func (slavemonitor *SlaveMonitoring) roundRobin() (int, error) {
	return 0, nil
}

func (slavemonitor *SlaveMonitoring) updateOneOfSlave(index int, status SlaveStatus, currentExecutingTask int) {
	slave := slavemonitor.SlavesAvailable[index]
	slavemonitor.SlavesAvailable[index] = Slave{
		ID:                  slave.ID,
		Address:             slave.Address,
		CurrentStatus:       status,
		CurrentExecuteTasks: append(slave.CurrentExecuteTasks, currentExecutingTask),
		Port:                slave.Port,
	}
}

func (slavemonitor *SlaveMonitoring) addNewTask(taskID string, slaveID int) {
	slavemonitor.CurrentTasks = append(slavemonitor.CurrentTasks, Task{
		ID:          taskID,
		TimeCreated: time.Now().Unix(),
		SlaveIndex:  slaveID,
	})
	slavemonitor.updateOneOfSlave(slaveID, SLAVE_WORKING, len(slavemonitor.CurrentTasks)-1)
}

func (slavemonitor *SlaveMonitoring) updateTaskStatus(taskID string, newStatus TaskStatus, timeFinished int64) error {
	for index, task := range slavemonitor.CurrentTasks {
		if task.ID == taskID {
			slavemonitor.CurrentTasks[index] = Task{
				ID:            task.ID,
				TimeCreated:   task.TimeCreated,
				TimeFinishing: timeFinished,
				SlaveIndex:    task.SlaveIndex,
				Status:        newStatus,
			}
			return nil
		}
	}
	return errors.New("can not update work status by undefined task")
}

/*GetTaskStatus - получить текущий статус задачи по её идентификатору*/
func (slavemonitor *SlaveMonitoring) GetTaskStatus(taskID string) (*TaskStatus, error) {
	for _, task := range slavemonitor.CurrentTasks {
		if task.ID == taskID {
			return &task.Status, nil
		}
	}
	return nil, errors.New("can not get task status by undefined task")
}
