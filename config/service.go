package config

import "github.com/goreflect/gostructor"

type ServiceConfig struct {
	APIPORT        int    `cf_env:"API_PORT" cf_default:"9999"`
	ConsulAddress  string `cf_env:"CONSUL_ADDRESS" cf_default:"127.0.0.1:8500"`
	ConsulUsername string `cf_env:"CONSUL_USERNAME" cf_default:"kubitre"`
	ConsulPassword string `cf_env:"CONSUL_PASSWORD" cf_default:"password"`
	ServiceType    string `cf_env:"SERVICE_TYPE" cf_default:"SLAVE"`     // MASTER, SLAVE
	ServicePlugin  string `cf_env:"SERVICE_PLUGIN" cf_default:"DEFAULT"` // DEFAULT, PORTAL
}

const (
	SERVICEMASTER = "MASTER"
	SERVICESLAVE  = "SLAVE"
)

const (
	PLUGINDEFAULT = "DEFAULT"
	PLUGINPORTAL  = "PORTAL"
)

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
