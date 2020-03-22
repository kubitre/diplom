package docker_runner

import (
	"testing"
	"time"

	"github.com/kubitre/diplom/models"
)

func Test_createDefaultContainer(t *testing.T) {
	dockerExecutor, errCreateExecutor := NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	if err := dockerExecutor.removeContainer("container_test_test2"); err != nil {
		t.Log("can not remove container. ", err.Error())
	}
	if err := dockerExecutor.PullImage("alpine"); err != nil {
		t.Error("can not pull alpine")
	}

	time.Sleep(time.Second * 5)

	if err := dockerExecutor.CreateContainer(&models.ContainerCreatePayload{
		BaseImageName: "alpine",
		WorkDir:       "/test",
		ShellCommands: []string{},
		ContainerName: "container_test_test2",
	}); err != nil {
		t.Error("can not create container. Error: ", err.Error())
	}
}

func Test_createContainerError(t *testing.T) {
	DockerExecutor, errCreate := NewDockerExecutor()
	if errCreate != nil {
		t.Error("can not create executor of docker. " + errCreate.Error())
	}
	if err := DockerExecutor.CreateContainer(&models.ContainerCreatePayload{}); err == nil {
		t.Error("can not create container")
	}
}
