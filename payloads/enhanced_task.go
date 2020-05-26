package payloads

import "github.com/kubitre/diplom/models"

type EnhancedSlave struct {
	ID                  string
	Address             string
	Port                int
	CurrentExecuteTasks []models.Task
}
