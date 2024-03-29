package docker_runner

import (
	"testing"

	"github.com/kubitre/diplom/gitmod"
	log "github.com/sirupsen/logrus"
)

// func Test_CreateImageDockerByDockerfile(t *testing.T) {
// 	dockerExecutor, errCreateExecutor := NewDockerExecutor()
// 	if errCreateExecutor != nil {
// 		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
// 	}
// 	log.SetLevel(log.DebugLevel)
// 	if err := dockerExecutor.CreateImageDockerFile("Dockerfile", []string{"test_candidate_1"}); err != nil {
// 		t.Error("can not create image by dockerfile. Error: ", err.Error())
// 	}
// }

func Test_CreateImageDockerByMemfile(t *testing.T) {
	dockerExecutor, errCreateExecutor := NewDockerExecutor()
	if errCreateExecutor != nil {
		t.Error("can not create client for docker manipulation. Error: ", errCreateExecutor.Error())
	}
	log.SetLevel(log.DebugLevel)
	gy := gitmod.Git{}
	repoPath, err := gy.CloneRepo("https://github.com/kubitre/for_diplom")
	if err != nil {
		t.Error(err)
	}
	if err := dockerExecutor.CreateImageMem([]string{
		"FROM golang:1.14.2-alpine3.11",
		"RUN apk update && apk add bash",
		"{{repoCandidate}}",
		"{{workdir repoCandidate}}",
		"RUN go build -o service",
	}, []string{
		`./service`,
	}, []string{"test_candidate_2"}, map[string]string{
		repoPath: "repoCandidate",
	}); err != nil {
		t.Error("can not create image by dockerfile. Error: ", err.Error())
	}
	if err := gy.RemoveRepo(repoPath); err != nil {
		t.Log("can not remove candidate repo. ", err)
	}
}
