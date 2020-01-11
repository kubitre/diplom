package docker

import (
	"testing"

	"github.com/kubitre/diplom/docker"
)

func Test_createDefaultContainer(t *testing.T) {
	dockerExecutor, errCreateExecutor := docker.NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	if err := dockerExecutor.CreateContainer(); err != nil {
		t.Error("can not create container. Error: ", err.Error())
	}
}
