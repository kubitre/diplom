package docker

import (
	"context"
	"os"

	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

type (
	DockerExecutor struct {
		Status       chan bool // syncronization for executor
		DockerClient *client.Client
	}
	EnvImage struct {
		Type  string `json:"type"`
		Image string `json:"payload"`
	}
)

// Initialize new docker client executor
func NewDockerExecutor() (*DockerExecutor, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Error("can not initiate client for docker api. Error: ", err.Error())
		return nil, err
	}
	return &DockerExecutor{
		Status:       make(chan bool, 1),
		DockerClient: cli,
	}, nil
}

// pulling image from hub.docker.com
func (docker *DockerExecutor) PullImage(image *EnvImage) error {
	ctx := context.Background()
	// resp, err := cli.NetworkCreate(ctx, "candidate_id_code", types.NetworkCreate{
	// 	Attachable: true,
	// 	Driver:     "bridge",
	// })
	// if err != nil {
	// 	log.Error("Can not create network docker. Error: ", err.Error())
	// 	return err
	// }
	// log.Debug("response from creating network: ", resp)
	respPulling, errPulling := docker.DockerClient.ImagePull(ctx, image.Image, types.ImagePullOptions{})
	if errPulling != nil {
		log.Error("Can npt pulling image. Error: ", errPulling.Error())
		return errPulling
	}
	// io.Copy(os.Stdout, respPulling)
	log.Debug("response from pulling image: ", respPulling)
	return nil
}

// CreateContainer - function for creating new container with docker file such as json
func (docker *DockerExecutor) CreateContainer() error {
	ctx := context.Background()
	repsCreating, err := docker.DockerClient.ContainerCreate(ctx, &container.Config{
		Image:      "ubuntu",
		WorkingDir: "/test",
		Shell: []string{
			"RUN ls -la",
		},
	}, &container.HostConfig{
		AutoRemove: true,
	}, &network.NetworkingConfig{}, "test_env_candidate_1")
	if err != nil {
		log.Error("can not create container with default configuration. Error: ", err.Error())
		return err
	}
	log.Debug("success create container. Output oprts: ", repsCreating)
	return nil
}

func (docker *DockerExecutor) CreateImage(dockerFilePath, buildContextPath string, tags []string) error {
	ctx := context.Background()
	buildOptions := types.ImageBuildOptions{
		Dockerfile: dockerFilePath,
		Tags:       tags,
	}

	buildContext, errArchive := archive.TarWithOptions(buildContextPath, &archive.TarOptions{})
	if errArchive != nil {
		log.Error("can not archive buildContextPath. Error: ", errArchive.Error())
		return errArchive
	}
	resp, err := docker.DockerClient.ImageBuild(ctx, buildContext, buildOptions)
	if err != nil {
		log.Error("error while build image by dockerfile. Error: ", err.Error())
		return err
	}
	log.Debug("response from building image: ", resp)
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil)
	return nil
}
