package docker

import (
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/kubitre/diplom/docker"
)

func Test_PullingHelloWorldImage(t *testing.T) {
	dockerExecutor, errCreateExecutor := docker.NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	log.SetLevel(log.DebugLevel)
	err := dockerExecutor.PullImage(&docker.EnvImage{
		Type:  "image",
		Image: "mcr.microsoft.com/azuredocs/aci-helloworld",
	})
	if err != nil {
		t.Error("can not pulling image. Error: ", err.Error())
	}
}
