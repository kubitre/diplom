package runner

import "github.com/kubitre/diplom/models"

type IRunner interface {
	Run(*models.Task) error
}
