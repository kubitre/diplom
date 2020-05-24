package routes

import "net/http"

type IMaster interface {
	CreateNewTask(http.ResponseWriter, *http.Request)
	ChangeTaskStatus(http.ResponseWriter, *http.Request)
	GetLogTask(http.ResponseWriter, *http.Request)
	CreateLogTask(http.ResponseWriter, *http.Request)
	GetTaskStatus(http.ResponseWriter, *http.Request)
	GetStatusWorkers(http.ResponseWriter, *http.Request)
	GetReportsPerTask(http.ResponseWriter, *http.Request)
}
