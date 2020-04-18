package payloads

import (
	"testing"

	"github.com/kubitre/diplom/masterexecutor/models"
	"gopkg.in/yaml.v2"
)

func TestMarshal(t *testing.T) {
	taskConfig := models.TaskConfig{
		TaskID: "testID",
		Stages: []string{
			"test1",
		},
		Jobs: map[string]models.Job{
			"Job1": models.Job{
					Stage: "test1",
					Image: []string{
						"FROM golang:1.14.2-alpine3.11",
						"RUN apk update && apk add bash",
						"{{repoCandidate}}",
						"{{workdir repoCandidate}}",
					},
					RepositoryCandidate: "https://github.com/kubitre/for_diplom.git",
					Reports: map[string]string{
						"allOutInfo": "^(?P<statusTest>FAIL|ok)\\s+(?P<Placement>[\\w_\\/]+)\\s+(?P<Time>[\\w.]+)$",
						"failedTest": "(?P<TEST>(--- FAIL: )(?P<TestName>[\\w]+)\\s+\\((?P<Time>[\\w.]+)\\))|(?P<Logs>\\s+(?P<fileName>[\\w_.]+):(?P<LineNumber>\\w+): (?P<LogText>.+))",
					},
			},
		},
	}
	result, _ := yaml.Marshal(&taskConfig)
	t.Log(result)
}