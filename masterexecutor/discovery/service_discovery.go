package discovery

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/kubitre/diplom/masterexecutor/config"
)

type Discovery struct {
	CurrentServiceName   string
	ConsulClient         *consulapi.Client
	ConfigurationService *config.ConfigurationRunner
}

/*InitializeDiscovery - инициализация текущего Discovery*/
func InitializeDiscovery(config *config.ConfigurationRunner) *Discovery {
	return &Discovery{
		CurrentServiceName:   "master-executor#" + uuid.New().String(),
		ConsulClient:         nil,
		ConfigurationService: config,
	}
}

/*NewClientForConsule - инициализация подключения до consul*/
func (discovery *Discovery) NewClientForConsule() error {
	config := consulapi.Config{
		Address: discovery.ConfigurationService.ConsulAddress,
		HttpAuth: &consulapi.HttpBasicAuth{
			Username: discovery.ConfigurationService.ConsulUsername,
			Password: discovery.ConfigurationService.ConsulPassword,
		},
	}
	consul, err := consulapi.NewClient(&config)
	if err != nil {
		return err
	}
	discovery.ConsulClient = consul
	return nil
}

/*RegisterServiceWithConsul - регистрация сервиса в consul*/
func (discovery *Discovery) RegisterServiceWithConsul() {
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = "master-executor#" + uuid.New().String()
	discovery.CurrentServiceName = registration.ID
	registration.Name = "master-executor"
	log.Println("registration information about out service: ", registration)
	address := hostname()
	registration.Address = address
	port, err := strconv.Atoi(port()[1:len(port())])
	if err != nil {
		log.Fatalln(err)
	}
	registration.Port = port
	registration.Check = new(consulapi.AgentServiceCheck)
	registration.Check.HTTP = fmt.Sprintf("http://%s:%v/health",
		address, port)
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	discovery.ConsulClient.Agent().ServiceRegister(registration)
}

/*UnregisterCurrentService - удаление сервиса из consul*/
func (discovery *Discovery) UnregisterCurrentService() {
	if err := discovery.ConsulClient.Agent().ServiceDeregister(discovery.CurrentServiceName); err != nil {
		log.Println(err)
	}
}

/*GetService - получение текущих сервисов из consul*/
func (discovery *Discovery) GetService(serviceName string) {

}

func port() string {
	p := os.Getenv("API_PORT")
	if len(strings.TrimSpace(p)) == 0 {
		return ":9997"
	}
	return fmt.Sprintf(":%s", p)
}

func hostname() string {
	hn, err := os.Hostname()
	if err != nil {
		log.Fatalln(err)
	}
	return hn
}
