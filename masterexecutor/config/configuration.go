package config

import "github.com/goreflect/gostructor"

/*ConfigurationRunner - все настройки по мастер ноде
 */
type ConfigurationRunner struct {
	APIPort        int    `cf_env:"API_PORT"`
	PathToLogsWork string `cf_env:"LOGS_WORK_PATH"`
}

/*ConfiureRunnerMaster - конфигурировании мастер ноды через Environment variables
 */
func ConfiureRunnerMaster() (*ConfigurationRunner, error) {
	struc, err := gostructor.ConfigureSmart(&ConfigurationRunner{}, "")
	if err != nil {
		return nil, err
	}
	return struc.(*ConfigurationRunner), nil
}
