package models

type (
	/*ContainerCreatePayload - payload for docker_runner module*/
	ContainerCreatePayload struct {
		BaseImageName string   `json:"image"`
		WorkDir       string   `json:"workdir"`
		ShellCommands []string `json:"shell"`
		ContainerName string   `json:"container_name"`
	}
)
