package payloads

type ChangeStatusJob struct {
	TaskID    string `json:"task_id"`
	NewStatus int    `json:"new_status"`
	Job       string `json:"job"`
}
