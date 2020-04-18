package models

type (
	TaskConfig struct {
		Jobs   map[string]Job `yaml:"jobs" json:"jobs"`
		Stages []string       `yaml:"stages" json:"stages"`
		TaskID string         `yaml:"taskID" taskID:"taskID"`
	}
)
