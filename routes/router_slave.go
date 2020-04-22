package routes

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/models"
	"gopkg.in/yaml.v2"
)

/*SlaveRunnerRouter - router for slave executor*/
type SlaveRunnerRouter struct {
	Router *mux.Router
	Core   *core.CoreSlaveRunner
}

const (
	apiTask        = "/task"
	apiHealthCheck = "/health"
)

/*InitNewSlaveRunnerRouter - initialize slave router*/
func InitNewSlaveRunnerRouter(core *core.CoreSlaveRunner) *SlaveRunnerRouter {
	return &SlaveRunnerRouter{
		Router: mux.NewRouter(),
		Core:   core,
	}
}

// createNewTask - создание новой задачи
func (route *SlaveRunnerRouter) createNewTask(writer http.ResponseWriter, request *http.Request) {
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

func (route *SlaveRunnerRouter) healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("service are running"))
}

// ConfigureRouter - конфигурирование роутера
func (route *SlaveRunnerRouter) ConfigureRouter() {
	log.Println("start configuring routes")
	route.Router.HandleFunc(apiTask, route.createNewTask).Methods(http.MethodPost)
	route.Router.HandleFunc(apiHealthCheck, route.healthCheck).Methods(http.MethodGet)
	log.Println("completed configuring routes")
}

// GetRouter - получение слейв роутера
func (route *SlaveRunnerRouter) GetRouter() *mux.Router {
	return route.Router
}
