package docker

import (
	"testing"

	"github.com/kubitre/diplom/docker"
	log "github.com/sirupsen/logrus"
)

func Test_createDefaultContainer(t *testing.T) {
	dockerExecutor, errCreateExecutor := docker.NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	log.SetLevel(log.DebugLevel)
	if err := dockerExecutor.CreateContainer(); err != nil {
		t.Error("can not create container. Error: ", err.Error())
	}
}
