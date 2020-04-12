package models

/*Job - primitive which parsed from entered yaml from portal*/
type Job struct {
	JobName             string   `yaml:"-"`
	Stage               string   `yaml:"stage"`
	TaskID              string   `yaml:"-"`
	Image               []string `yaml:"image"`
	RepositoryCandidate string   `yaml:"repo"`
	ShellCommands       []string `yaml:"run"`
}
