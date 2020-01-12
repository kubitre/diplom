package core

import (
	"errors"

	"github.com/kubitre/diplom/docker"
	"github.com/kubitre/diplom/gitmod"
	"github.com/kubitre/diplom/models"
	log "github.com/sirupsen/logrus"
)

type (
	CoreRunner struct {
		Git        *gitmod.Git
		Docker     *docker.DockerExecutor
		WorkerPull chan bool
		WorkConfig *models.WorkConfig
	}
	Worker struct {
		Task   chan models.Task
		Result chan models.OutputTask
	}
)

func NewCoreRunner(
	git *gitmod.Git,
	dock *docker.DockerExecutor,
	amountWorkers int,
	runnerConfig *models.WorkConfig) (*CoreRunner, error) {
	if dock == nil {
		return nil, errors.New("can not create runner without docker client")
	}
	return &CoreRunner{
		Git:        git,
		Docker:     dock,
		WorkerPull: make(chan bool, amountWorkers),
		WorkConfig: runnerConfig,
	}, nil
}

/*SetupConfiguration - setting up configuration if its no configuring in start*/
func (core *CoreRunner) SetupConfigurationPipeline(config *models.WorkConfig) error {
	if len(config.Stages) == 0 {
		return errors.New("runner config should have 1 or mode stages annotation")
	}
	if len(config.Tasks) == 0 {
		return errors.New("runner config should have 1 or more tasks")
	}
	core.WorkConfig = config
	return nil
}

func (core *CoreRunner) CreatePipeline(config *models.WorkConfig) error {
	if core.WorkConfig == nil {
		if config == nil {
			return errors.New("can not create pipeline without configuration. Please setup configuration and continue")
		}
		core.WorkConfig = config
	}
	for _, stage := range config.Stages {
		log.Info("start working on stage: " + stage)
		if err := core.executingTaskInStage(stage); err != nil {
			log.Error("can not execute stage. ", err)
			return err
		}
	}
	return nil
}

func (core *CoreRunner) executingTaskInStage(stage string) error {
	currentTasks := core.getTasksByStage(stage)
	for taskIdName, task := range currentTasks {
		log.Info("current task: ", task)
		if err := core.prepareTask(taskIdName, &task); err != nil {
			log.Error("error while preparing task. ", err)
			return err
		}
	}
	return nil
}

func (core *CoreRunner) getRepoCandidate(taskName string) error {
	if err := core.Git.CloneRepo(core.WorkConfig.Tasks[taskName].RepositoryCandidate); err != nil {
		return err
	}
	return nil
}

func (core *CoreRunner) prepareTask(taskName string, task *models.Task) error {
	if err := core.getRepoCandidate(taskName); err != nil {
		return err
	}
	if err := core.Docker.CreateImageMem(task.Image, []string{core.WorkConfig.RunID + taskName}); err != nil {
		return err
	}
	return nil
}

func (core *CoreRunner) getTasksByStage(stage string) map[string]models.Task {
	result := make(map[string]models.Task)
	for taskId, task := range core.WorkConfig.Tasks {
		log.Info("current task id: ", taskId)
		if task.Stage == stage {
			result[taskId] = task
		}
	}
	return result
}
