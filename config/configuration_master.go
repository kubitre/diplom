package config

import "github.com/goreflect/gostructor"

/*ConfigurationMasterRunner - все настройки по мастер ноде
 */
type ConfigurationMasterRunner struct {
	PathToLogsWork  string `cf_env:"LOGS_WORK_PATH"`
	MaxTaskPerSlave int    `cf_env:"MAX_TASKS_PER_SLAVE"`
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
