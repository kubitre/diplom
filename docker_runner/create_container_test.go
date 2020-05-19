package docker_runner

import (
	"bytes"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
)

func Test_createDefaultContainer(t *testing.T) {
	dockerExecutor, errCreateExecutor := NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
		return
	}
	if err := dockerExecutor.RemoveContainer("container_test_test10"); err != nil {
		t.Log("can not remove container. ", err.Error())
	}

	containerID, err := dockerExecutor.CreateContainer(&models.ContainerCreatePayload{
		BaseImageName: "test_candidate_2:latest",
		ContainerName: "container_test_test10",
	})
	if err != nil {
		t.Error("can not create container. Error: ", err.Error())
	}

	respCloser, errStart := dockerExecutor.RunContainer(containerID)
	if errStart != nil {
		t.Error("can not start container: ", errStart)
		return
	}
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	writer, err := stdcopy.StdCopy(stdout, stderr, respCloser)
	if err != nil {
		panic(err)
	}
	t.Log("logs from container: ", writer)
}
