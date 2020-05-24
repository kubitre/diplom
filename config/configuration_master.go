package config

import "github.com/goreflect/gostructor"

/*ConfigurationMasterRunner - все настройки по мастер ноде
 */
type ConfigurationMasterRunner struct {
	PathToLogsWork    string `cf_env:"LOGS_WORK_PATH" cf_default:"logs"`
	PathToReportsWork string `cf_env:"REPORT_WORK_PATH" cf_default:"reports"`
	MaxTaskPerSlave   int    `cf_env:"MAX_TASKS_PER_SLAVE" cf_env:"10"`
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
