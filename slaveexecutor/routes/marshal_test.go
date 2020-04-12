package routes

import (
	"testing"

	"github.com/kubitre/diplom/slaveexecutor/models"
	"gopkg.in/yaml.v2"
)

func TestMarshaling(t *testing.T) {
	model := models.TaskConfig{
		TaskID: "1234CF",
		Stages: []string{
			"build",
			"test",
			"lint",
		},
		Jobs: map[string]models.Job{
			"MyFirstJob": models.Job{
				Stage: "build",
				Image: []string{
					"FROM golang:1.14.2-alpine3.11 as builder",
				},
				ShellCommands: []string{
					"go build ",
					"./service",
				},
				RepositoryCandidate: "https://github.com/kubitre/for_diplom",
			},
		},
	}
	marshaled, errMarshal := yaml.Marshal(model)
	if errMarshal != nil {
		t.Error(errMarshal)
	}
	t.Error(marshaled)
}
