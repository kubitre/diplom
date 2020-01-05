package docker

import (
	"testing"

	"github.com/kubitre/diplom/docker"
)

func Test_CreateImageDocker(t *testing.T) {
	dockerExecutor, errCreateExecutor := docker.NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	log.SetLevel(log.DebugLevel)
	if err := dockerExecutor.CreateImage("", "", []string{}); err != nil {
		t.Error("can not create image by dockerfile. Error: ", err.Error())
	}
}
