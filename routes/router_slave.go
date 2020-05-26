package routes

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/core"
	"github.com/kubitre/diplom/models"
	log "github.com/sirupsen/logrus"
)

/*SlaveRunnerRouter - router for slave executor*/
type SlaveRunnerRouter struct {
	Router *mux.Router
	Core   *core.SlaveRunnerCore
}

/*InitNewSlaveRunnerRouter - initialize slave router*/
func InitNewSlaveRunnerRouter(core *core.SlaveRunnerCore) *SlaveRunnerRouter {
	return &SlaveRunnerRouter{
		Router: mux.NewRouter(),
		Core:   core,
	}
}

// createNewTask - создание новой задачи
func (route *SlaveRunnerRouter) createNewTask(writer http.ResponseWriter, request *http.Request) {
	var model models.TaskConfig
	if errDecode := json.NewDecoder(request.Body).Decode(&model); errDecode != nil {
		log.Println("can not parsed input task: ", errDecode)
		writer.WriteHeader(http.StatusBadRequest)
		return
	}
	log.Println("start executing new task: ", model)
	go func() {
		if errCreateTask := route.Core.CreatePipeline(&model); errCreateTask != nil {
			log.Error("can not create this")
			return
		}
	}()
	log.Println("completed prepared for task: ", model.TaskID)
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("completed saved and start preparing task for working with that"))
}

func (route *SlaveRunnerRouter) healthCheck(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte("service are running"))
}

// ConfigureRouter - конфигурирование роутера
func (route *SlaveRunnerRouter) ConfigureRouter() {
	log.Println("start configuring routes")
	route.Router.HandleFunc(ApiTask, route.createNewTask).Methods(http.MethodPost)
	route.Router.HandleFunc(ApiHealthCheck, route.healthCheck).Methods(http.MethodGet)
	log.Println("completed configuring routes")
}

// GetRouter - получение слейв роутера
func (route *SlaveRunnerRouter) GetRouter() *mux.Router {
	return route.Router
}
