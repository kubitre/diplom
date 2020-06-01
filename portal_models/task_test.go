package portal_models

import (
	"testing"

	"github.com/kubitre/diplom/models"
	"github.com/stretchr/testify/assert"
)

func Test_ConvertFromPortalObjectToAgentTask(t *testing.T) {
	modelPortal := PortalTask{
		TaskID: "test",
		JobGroups: []JobGroup{
			JobGroup{
				NameGroup: "stage1",
				Order:     1,
				Jobs: []Job{
					Job{
						JobName:    "test1",
						Dockerfile: "FROM ubuntu\nRUN apt update && apt install\nCMD echo \"kubitre awesome\"",
						Timeout:    "19:00:00",
						Metrics: []Metric{
							Metric{
								MetricName: "checkTest",
								Regex:      "mySuperRegex",
							},
						},
					},
				},
			},
		},
	}
	agentTask := modelPortal.ConvertToAgentTask()
	t.Log(agentTask)
	assert.Equal(t, models.TaskConfig{
		TaskID: "test",
		Stages: []string{
			"stage1",
		},
		Jobs: map[string]models.Job{
			"test1": models.Job{
				Image: []string{
					"FROM ubuntu",
					"RUN apt update && apt install",
					"CMD echo \"kubitre awesome\"",
				},
				Stage: "stage1",
				Reports: map[string]string{
					"checkTest": "mySuperRegex",
				},
				TaskID:  "test",
				JobName: "test1",
			},
		},
	}, agentTask, nil)
}
