package routes

import (
	"net/http"

	"github.com/gorilla/mux"
)

/*IMaster - интерфейс, который должны реализовать любые плагины для мастер ноды*/
type IMaster interface {
	CreateNewTask(http.ResponseWriter, *http.Request)
	ChangeTaskStatus(http.ResponseWriter, *http.Request)
	GetLogTask(http.ResponseWriter, *http.Request)
	CreateLogTask(http.ResponseWriter, *http.Request)
	GetTaskStatus(http.ResponseWriter, *http.Request)
	GetStatusWorkers(http.ResponseWriter, *http.Request)
	GetReportsPerTask(http.ResponseWriter, *http.Request)
	GetRouter() *mux.Router // system method
	ConfigureRouter()       //system method
}
