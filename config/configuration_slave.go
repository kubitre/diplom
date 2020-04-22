package config

import "github.com/goreflect/gostructor"

/*ConfigurationSlaveRunner - конфигурация слейв ноды
 */
type ConfigurationSlaveRunner struct {
	APIPORT                    int    `cf_env:"API_PORT"`
	ConsulAddress              string `cf_env:"CONSUL_ADDRESS"`
	ConsulUsername             string `cf_env:"CONSUL_USERNAME"`
	ConsulPassword             string `cf_env:"CONSUL_PASSWORD"`
	AmountPullWorkers          int    `cf_env:"AMOUNT_PULL_WORKERS"`
	AmountParallelTaskPerStage int    `cf_env:"AMOUNT_PARALLEL_TASK_PER_STAGE"`
}

/*ConfigureService - конфигурирования slave сервиса
 */
func ConfigureService() (*ConfigurationSlaveRunner, error) {
	config, errConfigure := gostructor.ConfigureSmart(&ConfigurationSlaveRunner{}, "")
	if errConfigure != nil {
		return nil, errConfigure
	}
	return config.(*ConfigurationSlaveRunner), nil
}

/*SetupNewPort - install port for binding*/
func (config *ConfigurationSlaveRunner) SetupNewPort(port int) {
	config.API_PORT = port
}
