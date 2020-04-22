package payloads

import "errors"

/*ChangeStatusTask - изменение текущего статуса для задачи проверки конкретного решения
*/
type ChangeStatusTask struct {
	TaskID       string `json:"work_id"`
	NewStatus    int    `json:"new_status"`
	Stage        string `json:"failed_stage"`
	TimeFinished int64  `json:"time_finished"`
}

/*Validate - валидация пришедшего обновления статуса*/
func (statusWork *ChangeStatusTask) Validate() error {
	if statusWork.NewStatus < 0 || statusWork.NewStatus > 5 {
		return errors.New("can not find this status")
	}
	return nil
}
