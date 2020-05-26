package config

import "github.com/goreflect/gostructor"

/*ConfigurationMasterRunner - все настройки по мастер ноде
 */
type ConfigurationMasterRunner struct {
	PathToLogsWork        string `cf_env:"LOGS_WORK_PATH" cf_default:"logs"`
	PathToReportsWork     string `cf_env:"REPORT_WORK_PATH" cf_default:"reports"`
	MaxTaskPerSlave       int    `cf_env:"MAX_TASKS_PER_SLAVE" cf_default:"10"`
	AgentID               string `cf_env:"AGENT_ID" cf_default:"default_agent"`
	AverageTimeoutPerTask int    `cf_env:"AVERAGE_TIMEOUT_PER_TASK"`
}

/*ConfiureRunnerMaster - конфигурировании мастер ноды через Environment variables
 */
func ConfiureRunnerMaster() (*ConfigurationMasterRunner, error) {
	struc, err := gostructor.ConfigureSmart(&ConfigurationMasterRunner{}, "")
	if err != nil {
		return nil, err
	}
	return struc.(*ConfigurationMasterRunner), nil
}
