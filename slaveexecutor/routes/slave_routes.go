package routes

import (
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
	apiTask = "/task"
)

func InitNewSlaveRouter() *SlaveRouter {
	return &SlaveRouter{
		Router: mux.NewRouter(),
	}
}

// createNewTask - создание новой задачи
func (route *SlaveRouter) createNewTask(writer http.ResponseWriter, request *http.Request) {
	var model models.TaskConfig
	if errDecode := yaml.NewDecoder(request.Body).Decode(&model); errDecode != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	if errCreateTask := route.Core.CreatePipeline(&model); errCreateTask != nil {
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	writer.WriteHeader(http.StatusOK)
}

// ConfigureRouter - конфигурирование роутера
func (route *SlaveRouter) ConfigureRouter() {
	route.Router.HandleFunc(apiTask, route.createNewTask).Methods(http.MethodPost)
}

// GetRouter - получение слейв роутера
func (route *SlaveRouter) GetRouter() *mux.Router {
	return route.Router
}
