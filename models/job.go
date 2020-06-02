package models

/*Job - primitive which parsed from entered yaml from portal*/
type Job struct {
	JobName             string            `yaml:"-"`
	Stage               string            `yaml:"stage" json:"stage"`
	TaskID              string            `yaml:"-"`
	Image               []string          `yaml:"image" json:"image"`
	Timeout             int64             `yaml:"timeout" json:"timeout"`
	RepositoryCandidate string            `yaml:"repo" json:"repo"`
	ShellCommands       []string          `yaml:"run" json:"run"`
	Reports             map[string]string `yaml:"reports" json:"reports"`
}
