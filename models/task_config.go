package models

type (
	/*TaskConfig - configuration task by description jobs, stages, identifier of task*/
	TaskConfig struct {
		Jobs   map[string]Job `yaml:"jobs"`
		Stages []string       `yaml:"stages"`
		TaskID string         `yaml:"taskID"`
	}
)
