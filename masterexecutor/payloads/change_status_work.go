package payloads

/*
ChangeStatusWork - изменение текущего статуса для задачи проверки конкретного решения
*/
type ChangeStatusWork struct {
	WorkID    string `json:"work_id"`
	NewStatus int    `json:"new_status"`
}
