package models

import "github.com/kubitre/diplom/tasks"

type RunnerConfig struct {
	Tasks  map[string]tasks.Task `yaml:"tasks"`
	Stages []string              `yaml:"stages"`
}
