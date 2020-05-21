package discovery

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/google/uuid"
	consulapi "github.com/hashicorp/consul/api"
	"github.com/kubitre/diplom/config"
)

const (
	slavePattern  = "slave-executor#"
	masterPattern = "master-executor#"
	tagSlave      = "slave"
	tagMaster     = "master"
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
		ConsulClient:       nil,
		ServiceConfig:      configService,
	}
}

/*NewClientForConsule - инициализация подключения до consul*/
func (discovery *Discovery) NewClientForConsule() error {
	log.Println("initialize new client for consul")
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
	log.Println("start registration" + discovery.CurrentServiceName + "in consul")
	registration := new(consulapi.AgentServiceRegistration)
	registration.ID = discovery.CurrentServiceName
	registration.Name = discovery.CurrentServiceType
	registration.Tags = tags
	log.Println("registration information about out service: ", registration)
	address := hostname()
	registration.Address = address
	registration.Port = discovery.ServiceConfig.APIPORT
	registration.Check = new(consulapi.AgentServiceCheck)
	registration.Check.HTTP = "http://" + address + ":" + strconv.Itoa(discovery.ServiceConfig.APIPORT) + "/health"
	registration.Check.Interval = "5s"
	registration.Check.Timeout = "3s"
	log.Println("registration information: ", registration.Check.HTTP)
	if errRegister := discovery.ConsulClient.Agent().ServiceRegister(registration); errRegister != nil {
		log.Println("can not registering in consule: ", errRegister)
		os.Exit(1)
	}
	log.Println("completed registered service in consul")
}

/*UnregisterCurrentService - удаление сервиса из consul*/
func (discovery *Discovery) UnregisterCurrentService() {
	log.Println("start de register service in consul")
	if err := discovery.ConsulClient.Agent().ServiceDeregister(discovery.CurrentServiceName); err != nil {
		log.Println(err)
	}
}

/*GetService - получение текущих сервисов из consul*/
func (discovery *Discovery) GetService(serviceName, tag string) []*consulapi.CatalogService {
	log.Println("getting service from consul by service name: ", serviceName)
	allServices, _, err := discovery.ConsulClient.Catalog().Service(serviceName, tag, nil)
	if err != nil {
		log.Println(err)
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
