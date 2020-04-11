package payloads

/*
CreateNewWork - создание новой задачи на проверку
*/
type CreateNewWork struct {
	WorkID string `json:"work_id"`
	Work   string `json:"work"` // string with yaml format
}
