package models

/*Task - primitive which parsed from entered yaml from portal*/
type Task struct {
	Stage               string   `yaml:"stage"`
	Image               []string `yaml:"image"`
	RepositoryCandidate string   `yaml:"repo"`
	ShellCommands       []string `yaml:"run"`
}
