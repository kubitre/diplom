package config

import "github.com/goreflect/gostructor"

type ServiceConfig struct {
	APIPORT        int    `cf_env:"API_PORT"`
	ConsulAddress  string `cf_env:"CONSUL_ADDRESS"`
	ConsulUsername string `cf_env:"CONSUL_USERNAME"`
	ConsulPassword string `cf_env:"CONSUL_PASSWORD"`
}

/*SetupNewPort - install port for binding*/
func (config *ServiceConfig) SetupNewPort(port int) {
	config.APIPORT = port
}

/*ConfigureService - конфигурирование общих настроек для сервисов
 */
func ConfigureService() (*ServiceConfig, error) {
	config, errConfigure := gostructor.ConfigureSmart(&ServiceConfig{}, "")
	if errConfigure != nil {
		return nil, errConfigure
	}
	return config.(*ServiceConfig), nil
}
