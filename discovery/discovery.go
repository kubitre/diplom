package discovery

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/kubitre/diplom/config"
	log "github.com/sirupsen/logrus"
)

const (
	SlavePattern  = "slave-executor#"
	MasterPattern = "master-executor#"
	TagSlave      = "slave"
	TagMaster     = "master"
)

type Discovery struct {
	CurrentServiceName string
	CurrentServiceType string
	ConsulClient       *consulapi.Client
	ServiceConfig      *config.ServiceConfig
}

/*InitializeDiscovery - инициализация текущего Discovery*/
func InitializeDiscovery(
	typeService string,
	configService *config.ServiceConfig) *Discovery {
	return &Discovery{
		CurrentServiceName: typeService + uuid.New().String(),
		CurrentServiceType: typeService,
		ConsulClient:       nil,
		ServiceConfig:      configService,
	}
}

/*NewClientForConsule - инициализация подключения до consul*/
func (discovery *Discovery) NewClientForConsule() error {
	log.Info("initialize new client for consul")
	config := consulapi.Config{
		Address: discovery.ServiceConfig.ConsulAddress,
		HttpAuth: &consulapi.HttpBasicAuth{
			Username: discovery.ServiceConfig.ConsulUsername,
			Password: discovery.ServiceConfig.ConsulPassword,
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
func (discovery *Discovery) RegisterServiceWithConsul(tags []string) {
	log.Info("start registration" + discovery.CurrentServiceName + "in consul")
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = discovery.CurrentServiceName
	registration.Name = discovery.CurrentServiceType
	registration.Tags = tags
	log.Info("registration information about out service: ", registration)
	address := hostname()
	registration.Address = address
	registration.Port = discovery.ServiceConfig.APIPORT
	registration.Check = new(consulapi.AgentServiceCheck)
	registration.Check.HTTP = "http://" + address + ":" + strconv.Itoa(discovery.ServiceConfig.APIPORT) + "/health"
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	log.Info("registration information: ", registration.Check.HTTP)
	if errRegister := discovery.ConsulClient.Agent().ServiceRegister(registration); errRegister != nil {
		log.Error("can not registering in consule: ", errRegister)
		os.Exit(1)
	}
	log.Info("completed registered service in consul")
}

/*UnregisterCurrentService - удаление сервиса из consul*/
func (discovery *Discovery) UnregisterCurrentService() {
	log.Info("start de register service in consul")
	if err := discovery.ConsulClient.Agent().ServiceDeregister(discovery.CurrentServiceName); err != nil {
		log.Error(err)
	}
}

/*GetService - получение текущих сервисов из consul*/
func (discovery *Discovery) GetService(serviceName, tag string) []*consulapi.ServiceEntry {
	log.Info("getting service from consul by service name: ", serviceName)
	allHealthServices, _, err2 := discovery.ConsulClient.Health().Service(serviceName, tag, true, nil)
	if err2 != nil {
		log.Error(err2)
	}
	if len(allHealthServices) > 0 {
		// catalogServices := discovery.getCatalogService(serviceName, tag)
		return allHealthServices
	}
	return nil
}

func (discovery *Discovery) getCatalogService(serviceName, tag string) []*consulapi.CatalogService {
	allServices, _, err := discovery.ConsulClient.Catalog().Service(serviceName, tag, nil)
	if err != nil {
		log.Error(err)
	}
	return allServices
}

func port(port int) string {
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
