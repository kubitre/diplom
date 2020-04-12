package models

/*Job - primitive which parsed from entered yaml from portal*/
type Job struct {
	Stage               string   `yaml:"stage"`
	WorkID              string   `yaml:"workid"`
	Image               []string `yaml:"image"`
	RepositoryCandidate string   `yaml:"repo"`
	ShellCommands       []string `yaml:"run"`
}
