package config

import "github.com/goreflect/gostructor"

/*ConfigurationSlaveRunner - конфигурация слейв ноды
 */
type ConfigurationSlaveRunner struct {
	AmountPullWorkers          int `cf_env:"AMOUNT_PULL_WORKERS" cf_default:"10"`
	AmountParallelTaskPerStage int `cf_env:"AMOUNT_PARALLEL_TASK_PER_STAGE" cf_default:"100"`
}

/*ConfigureRunnerSlave - конфигурирования slave сервиса
 */
func ConfigureRunnerSlave() (*ConfigurationSlaveRunner, error) {
	config, errConfigure := gostructor.ConfigureSmart(&ConfigurationSlaveRunner{}, "")
	if errConfigure != nil {
		return nil, errConfigure
	}
	return config.(*ConfigurationSlaveRunner), nil
}
