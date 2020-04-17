package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/slaveexecutor/core"
	"github.com/kubitre/diplom/slaveexecutor/models"
	"gopkg.in/yaml.v2"
)

type SlaveRouter struct {
	Router *mux.Router
	Core   *core.CoreSlaveRunner
}

const (
	apiTask        = "/task"
	apiHealthCheck = "/health"
)

func InitNewSlaveRouter(core *core.CoreSlaveRunner) *SlaveRouter {
	return &SlaveRouter{
		Router: mux.NewRouter(),
		Core:   core,
	}
}

// createNewTask - создание новой задачи
func (route *SlaveRouter) createNewTask(writer http.ResponseWriter, request *http.Request) {
	var model models.TaskConfig
	if errDecode := yaml.NewDecoder(request.Body).Decode(&model); errDecode != nil {
		log.Println("can not parsed input task: ", errDecode)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("start executing new task: ", model.TaskID)
	if errCreateTask := route.Core.CreatePipeline(&model); errCreateTask != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("completed prepared for task: ", model.TaskID)
	writer.WriteHeader(http.StatusOK)
}

func (route *SlaveRouter) healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
}

// ConfigureRouter - конфигурирование роутера
func (route *SlaveRouter) ConfigureRouter() {
	log.Println("start configuring routes")
	route.Router.HandleFunc(apiTask, route.createNewTask).Methods(http.MethodPost)
	route.Router.HandleFunc(apiHealthCheck, route.healthCheck).Methods(http.MethodGet)
	log.Println("completed configuring routes")
}

// GetRouter - получение слейв роутера
func (route *SlaveRouter) GetRouter() *mux.Router {
	return route.Router
}
