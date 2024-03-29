package models

import (
	"encoding/json"
	"log"
)

type (
	/*TaskConfig - configuration task by description jobs, stages, identifier of task*/
	TaskConfig struct {
		Jobs   map[string]Job `yaml:"jobs" json:"jobs"`
		Stages []string       `yaml:"stages" json:"stages"`
		TaskID string         `yaml:"taskID" json:"taskID"`
	}
)

// Validate - валидация входящего задания в исполняющий модуль
func (task *TaskConfig) Validate() bool {
	return true
}

// ToByteArray - конвертация текущей модели в массив байтов для передачи по сети
func (task *TaskConfig) ToByteArray() ([]byte, error) {
	bts, err := json.Marshal(task)
	log.Println(string(bts))
	if err != nil {
		return nil, err
	}
	var model TaskConfig
	json.Unmarshal(bts, &model)
	log.Println("after unmarshaling: ", model)
	return bts, nil
}
