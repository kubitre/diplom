package portal_models

/*PortalTaskStatus - статус по задаче в формате портала*/
type PortalTaskStatus struct {
	TaskID             string            `json:"id"`
	TaskStatus         string            `json:"state"`
	UserViewResultData map[string]string `json:"data"`
	DeveloperOnlyData  map[string]string `json:"runner_data"`
}
