package config

import "github.com/goreflect/gostructor"

/*SlaveConfiguration - конфигурация слейв ноды
 */
type SlaveConfiguration struct {
	APIPort                    int    `cf_env:"SLAVE_API_PORT"`
	ConsulAddress              string `cf_env:"CONSUL_ADDRESS"`
	ConsulUsername             string `cf_env:"CONSUL_USERNAME"`
	ConsulPassword             string `cf_env:"CONSUL_PASSWORD"`
	AmountPullWorkers          int    `cf_env:"AMOUNT_PULL_WORKERS"`
	AmountParallelTaskPerStage int    `cf_env:"AMOUNT_PARALLEL_TASK_PER_STAGE"`
}

/*ConfigureService - конфигурирования slave сервиса
 */
func ConfiguringService() (*SlaveConfiguration, error) {
	config, errConfigure := gostructor.ConfigureSmart(&SlaveConfiguration{}, "")
	if errConfigure != nil {
		return nil, errConfigure
	}
	return config.(*SlaveConfiguration), nil
}
