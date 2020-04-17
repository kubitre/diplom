package models

type (
	TaskConfig struct {
		Jobs   map[string]Job `yaml:"jobs"`
		Stages []string       `yaml:"stages"`
		TaskID string         `yaml:"taskID"`
	}
)
