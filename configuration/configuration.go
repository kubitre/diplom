package configuration

type ConfigurationExecutor struct {
	WorkerAmoun   int    `valid:"notANumberValidator" default:"10"`
	PortalAddress string `valid:"notAStringValidator" default:"http://localhost:9999" environment:"portal"`
}
