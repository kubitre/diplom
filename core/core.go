package core

import (
	"github.com/kubitre/diplom/docker"
	"github.com/kubitre/diplom/gitmod"
)

type CoreRunner struct {
	Git            *gitmod.Git
	Docker *docker.DockerExecutor
}
