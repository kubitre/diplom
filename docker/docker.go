package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"io/ioutil"
	"os"

	"github.com/kubitre/diplom/models"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

type (
	DockerExecutor struct {
		Status       chan bool // syncronization for executor
		DockerClient *client.Client
	}
)

// Initialize new docker client executor
func NewDockerExecutor() (*DockerExecutor, error) {
	cli, err := client.NewClientWithOpts()
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
func (docker *DockerExecutor) PullImage(image string) error {
	ctx := context.Background()
	respPulling, errPulling := docker.DockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
	if errPulling != nil {
		log.Error("Can npt pulling image. Error: ", errPulling.Error())
		return errPulling
	}
	// io.Copy(os.Stdout, respPulling)
	log.Debug("response from pulling image: ", respPulling)
	return nil
}

// CreateContainer - function for creating new container with docker file such as json
func (docker *DockerExecutor) CreateContainer(payload *models.ContainerCreatePayload) error {
	ctx := context.Background()
	repsCreating, err := docker.DockerClient.ContainerCreate(ctx, &container.Config{
		Image:      payload.BaseImageName,
		WorkingDir: payload.WorkDir,
		Shell:      payload.ShellCommands,
	}, &container.HostConfig{
		AutoRemove: true,
	}, &network.NetworkingConfig{}, payload.ContainerName)
	if err != nil {
		log.Error("can not create container with default configuration. Error: ", err.Error())
		return docker.removeContainer("test_env_candidate_1")
	}
	log.Debug("success create container. Output oprts: ", repsCreating)
	return nil
}

func (docker *DockerExecutor) initiateTarFromFS(dockerFilePath string) (*bytes.Buffer, error) {
	dockerFileReader, err := os.Open(dockerFilePath)
	if err != nil {
		log.Error("can not open docker file for reading that. " + err.Error())
		return nil, err
	}
	readedDockerFile, err := ioutil.ReadAll(dockerFileReader)
	if err != nil {
		log.Error("can not read docker file. " + err.Error())
		return nil, err
	}
	return docker.tarCreate(dockerFilePath, readedDockerFile)
}

func (docker *DockerExecutor) tarCreate(filePath string, data []byte) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	tw := tar.NewWriter(buf)
	defer tw.Close()

	tarHeader := &tar.Header{
		Name: filePath,
		Size: int64(len(data)),
	}
	if err := tw.WriteHeader(tarHeader); err != nil {
		log.Error("can not write tar header. " + err.Error())
		return nil, err
	}
	_, err := tw.Write(data)
	if err != nil {
		log.Error("can not writing dockerfile into tar archive. " + err.Error())
		return nil, err
	}
	return buf, nil
}

func (docker *DockerExecutor) inituateTarFromStringArray(dockerFile []string) (*bytes.Buffer, error) {
	result := ""
	for _, v := range dockerFile {
		result += v + "\n"
	}

	return docker.tarCreate("Dockerfile", []byte(result))
}

func (docker *DockerExecutor) CreateImageMem(dockerFile, tags []string) error {
	ctx := context.Background()
	resultBuffer, err := docker.inituateTarFromStringArray(dockerFile)
	if err != nil {
		log.Error("can not readed bytes from fs. " + err.Error())
		return err
	}

	dockerFileTar := bytes.NewReader(resultBuffer.Bytes())
	buildOptions := types.ImageBuildOptions{
		Context:    dockerFileTar,
		Dockerfile: "Dockerfile",
		Tags:       tags,
	}
	resp, err := docker.DockerClient.ImageBuild(ctx, dockerFileTar, buildOptions)
	if err != nil {
		log.Error("error while build image by dockerfile. Error: ", err.Error())
		return err
	}
	log.Debug("response from building image: ", resp)
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	if err1 := jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil); err1 != nil {
		return err1
	}

	return nil
}

func (docker *DockerExecutor) CreateImageDockerFile(dockerFilePath string, tags []string) error {
	ctx := context.Background()
	resultBuffer, err := docker.initiateTarFromFS(dockerFilePath)
	if err != nil {
		log.Error("can not readed bytes from fs. " + err.Error())
		return err
	}

	dockerFileTar := bytes.NewReader(resultBuffer.Bytes())

	buildOptions := types.ImageBuildOptions{
		Context:    dockerFileTar,
		Dockerfile: dockerFilePath,
		Tags:       tags,
	}
	resp, err := docker.DockerClient.ImageBuild(ctx, dockerFileTar, buildOptions)
	if err != nil {
		log.Error("error while build image by dockerfile. Error: ", err.Error())
		return err
	}
	log.Debug("response from building image: ", resp)
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	if err1 := jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil); err1 != nil {
		return err1
	}
	return nil
}

func (docker *DockerExecutor) removeContainer(containerName string) error {
	ctx := context.Background()
	if err := docker.DockerClient.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{}); err != nil {
		log.Error("can not remove container. " + err.Error())
		return err
	}
	return nil
}
