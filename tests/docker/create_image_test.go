package docker

import (
	"testing"

	"github.com/kubitre/diplom/docker"
	log "github.com/sirupsen/logrus"
)

func Test_CreateImageDockerByDockerfile(t *testing.T) {
	dockerExecutor, errCreateExecutor := docker.NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	log.SetLevel(log.DebugLevel)
	if err := dockerExecutor.CreateImageDockerFile("Dockerfile", []string{"test_candidate_1"}); err != nil {
		t.Error("can not create image by dockerfile. Error: ", err.Error())
	}
}

func Test_CreateImageDockerByMemfile(t *testing.T) {
	dockerExecutor, errCreateExecutor := docker.NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	log.SetLevel(log.DebugLevel)
	if err := dockerExecutor.CreateImageMem([]string{
		"FROM ubuntu:18.04",
		"RUN apt update",
	}, []string{"test_candidate_2"}); err != nil {
		t.Error("can not create image by dockerfile. Error: ", err.Error())
	}
}
