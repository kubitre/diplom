package docker

import (
	"testing"
)

func Test_createDefaultContainer(t *testing.T) {
	dockerExecutor, errCreateExecutor := NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	if err := dockerExecutor.CreateContainer(); err != nil {
		t.Error("can not create container. Error: ", err.Error())
	}
}
