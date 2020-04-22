package payloads

import (
	"log"

	"github.com/kubitre/diplom/masterexecutor/models"
	"gopkg.in/yaml.v2"
)

/*CreateNewTask - создание новой задачи на проверку
*/
type CreateNewTask struct {
	TaskID string `json:"task_id"`
	Task   []byte `json:"task"` // string with yaml format
}

/*ConvertToTaskConfigBytes - convert from string to TaskConfig
 */
func (payload *CreateNewTask) ConvertToTaskConfigBytes() ([]byte, error) {
	var result models.TaskConfig

	log.Println("input payload: ", payload.Task)

	if err := yaml.Unmarshal(payload.Task, &result); err != nil {
		log.Println("can not unmarshal by yaml into TaskConfig: ", err)
		return nil, err
	}
	return payload.Task, nil
}
