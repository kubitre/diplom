package models

type (
	WorkConfig struct {
		Tasks  map[string]Task `yaml:"tasks"`
		Stages []string        `yaml:"stages"`
		RunID  string          `yaml:"workID"`
	}
)
