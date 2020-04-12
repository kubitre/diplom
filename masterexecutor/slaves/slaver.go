package slaves

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/kubitre/diplom/masterexecutor/payloads"
)

type (
	SlaveMonitoring struct {
		SlavesAvailable          []Slave
		LastUsingService         chan int
		CurrentTasks             []Task
		MaxExecutingTaskPerSlave int
	}

	Task struct {
		ID            string
		SlaveIndex    int
		Status        TaskStatus
		TimeCreated   int64
		TimeFinishing int64
	}

	Slave struct {
		ID                  string
		Address             string
		Port                int
		CurrentStatus       SlaveStatus
		CurrentExecuteTasks []int // index of SlaveMonitoring.CurrentTasks
	}

	TaskStatus struct {
		StatusIndex TaskStatusIndx
		Stage       string
	}

	TaskStatusIndx int
	SlaveStatus    int
)

const (
	NEW      TaskStatusIndx = 0 // task saved in portal
	QUEUED   TaskStatusIndx = 1
	RUNNING                 = 2
	CANCELED                = 3
	FAILED                  = 4
	SUCCESS                 = 5
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
			slavemonitor.SlavesAvailable = append(slavemonitor.SlavesAvailable, Slave{
				ID:                  value.ID,
				Address:             value.Address,
				Port:                value.ServicePort,
				CurrentStatus:       WAITING_WORK,
				CurrentExecuteTasks: []int{},
			})
		}
	}
}

func (slavemonitor *SlaveMonitoring) notExistService(service *consulapi.CatalogService) bool {
	for _, slave := range slavemonitor.SlavesAvailable {
		if slave.ID == service.ID {
			return false
		}
	}
	return true
}

func (slavemonitor *SlaveMonitoring) changeLastUsingServiceIndex(newindex int) {
	slavemonitor.LastUsingService <- newindex
}

// SendSlaveWork - проксирование запроса от клиента на один из слейв сервисов
func (slavemonitor *SlaveMonitoring) SendSlaveWork(request *http.Request, writer http.ResponseWriter) error {
	vars := mux.Vars(request)
	taskID := vars["taskid"]
	if taskID == "" {
		return errors.New("value of taskID can not be null or empty")
	}
	slaveID, err := slavemonitor.chooseHaveSpaceForWorkSlave()
	if err != nil {
		return err
	}
	slavemonitor.addNewTask(taskID, slaveID)
	http.Redirect(writer, request, "http://"+slavemonitor.SlavesAvailable[slaveID].Address+":"+strconv.Itoa(slavemonitor.SlavesAvailable[slaveID].Port), http.StatusUseProxy)
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
	if len(slavemonitor.SlavesAvailable)-1 > lastUsedIndex {
		currentUseSlaveIndex = lastUsedIndex + 1
		if len(slavemonitor.SlavesAvailable[currentUseSlaveIndex].CurrentExecuteTasks) < slavemonitor.MaxExecutingTaskPerSlave {
			slavemonitor.changeLastUsingServiceIndex(currentUseSlaveIndex)
			return currentUseSlaveIndex, nil
		}
		for index, slave := range slavemonitor.SlavesAvailable {
			if len(slave.CurrentExecuteTasks) < slavemonitor.MaxExecutingTaskPerSlave {
				slavemonitor.changeLastUsingServiceIndex(index)
				return index, nil
			}
		}
	}
	return 0, errors.New("can not choose worker slave for executing this task")
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
