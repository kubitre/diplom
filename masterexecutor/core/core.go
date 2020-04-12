package core

import (
	"github.com/gorilla/mux"
	"github.com/kubitre/diplom/masterexecutor/config"
	"github.com/kubitre/diplom/masterexecutor/discovery"
	"github.com/kubitre/diplom/masterexecutor/routes"
	"github.com/kubitre/diplom/masterexecutor/slaves"
)

/*RunnerCore - ядро master ноды*/
type RunnerCore struct {
	RouterRunner  *routes.RunnerRouter
	Discovery     *discovery.Discovery
	SlaveMoniring *slaves.SlaveMonitoring
}

/*InitNewCore - инициализация ядра текущего сервиса*/
func InitNewCore(config *config.ConfigurationRunner) *RunnerCore {
	return &RunnerCore{
		RouterRunner: routes.InitializeRunnerRouter(),
		Discovery:    discovery.InitializeDiscovery(config),
	}
}

/*Run - запуск роутера, discovery, получение информации о слейвах*/
func (core *RunnerCore) Run(config *config.ConfigurationRunner) {
	core.Discovery.NewClientForConsule()
	core.Discovery.RegisterServiceWithConsul()
	core.RouterRunner.ConfiguringRoutes()
}

/*UnregisterService - де регистрация сервиса из consul*/
func (core *RunnerCore) UnregisterService() {
	core.Discovery.UnregisterCurrentService()
}

/*GetRouter - получение сконфигурированного роутера*/
func (core *RunnerCore) GetRouter() *mux.Router {
	return core.RouterRunner.GetRouterMux()
}
