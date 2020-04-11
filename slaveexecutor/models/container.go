package models

type (
	ContainerCreatePayload struct {
		BaseImageName string   `json:"image"`
		WorkDir       string   `json:"workdir"`
		ShellCommands []string `json:"shell"`
		ContainerName string   `json:"container_name"`
	}
)
