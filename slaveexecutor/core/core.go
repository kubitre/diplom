package core

import (
	"errors"
	"os"

	"github.com/kubitre/diplom/slaveexecutor/docker_runner"
	"github.com/kubitre/diplom/slaveexecutor/gitmod"
	"github.com/kubitre/diplom/slaveexecutor/models"
	log "github.com/sirupsen/logrus"
)

type (
	/*CoreSlaveRunner - ядро для слейва*/
	CoreSlaveRunner struct {
		Git        *gitmod.Git
		Docker     *docker_runner.DockerExecutor
		WorkerPull chan bool
		WorkConfig *models.WorkConfig
	}
	/*Worker - единичная воркер функция, которая отвечает за выполнение всех job на одной стадии одной задачи*/
	Worker struct {
		Task   chan models.Task       // текущая задача у воркера
		Result chan models.OutputTask // текущий результат у воркера (лог файл и попытка спарсить результат)
	}
)

/*NewCoreSlaveRunner - инициализация нового ядра слейв модуля*/
func NewCoreSlaveRunner(
	amountWorkers int,
	runnerConfig *models.WorkConfig) (*CoreRunner, error) {
	dock, err := docker_runner.NewDockerExecutor()
	if err != nil {
		log.Error("can not create docker executor. " + err.Error())
	}
	return &CoreRunner{
		Git:        &gitmod.Git{},
		Docker:     dock,
		WorkerPull: make(chan bool, amountWorkers),
		WorkConfig: runnerConfig,
	}, nil
}

/*SetupConfigurationPipeline - setting up configuration if its no configuring in start*/
func (core *CoreSlaveRunner) SetupConfigurationPipeline(config *models.WorkConfig) error {
	if len(config.Stages) == 0 {
		return errors.New("runner config should have 1 or mode stages annotation")
	}
	if len(config.Tasks) == 0 {
		return errors.New("runner config should have 1 or more tasks")
	}
	core.WorkConfig = config
	return nil
}

/*CreatePipeline - создание пайплайна на выполнение одной задачи*/
func (core *CoreSlaveRunner) CreatePipeline(config *models.WorkConfig) error {
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

func (core *CoreSlaveRunner) executingTaskInStage(stage string) error {
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

func (core *CoreSlaveRunner) getRepoCandidate(taskName string) (string, error) {
	path, err := core.Git.CloneRepo(core.WorkConfig.Tasks[taskName].RepositoryCandidate)
	if err != nil {
		switch core.Git.GetTypeError(err) {
		case gitmod.ErrorAuthenticate:
			return "", err
		case gitmod.ErrorExistingRepository:
			return path, core.Git.RemoveRepo(path)
		case gitmod.ErrorUnrecognized:
			return "", err
		}
	}
	return path, nil
}

func (core *CoreSlaveRunner) prepareTask(taskName string, task *models.Task) error {
	pathRepo, err := core.getRepoCandidate(taskName)
	if err != nil {
		return err
	}

	if err := core.Docker.CreateImageMem(core.appendRepoIntoDocker(pathRepo, task),
		task.ShellCommands,
		[]string{core.WorkConfig.RunID + taskName},
		map[string]string{}); err != nil {
		return err
	}
	if err := os.RemoveAll(pathRepo); err != nil {
		log.Warn("can not remove repo candidate. ", err)
	}
	return nil
}

func (core *CoreSlaveRunner) appendRepoIntoDocker(path string, task *models.Task) []string {
	return append(task.Image, `COPY `+path+` /repoCandidate`)
}

func (core *CoreSlaveRunner) getTasksByStage(stage string) map[string]models.Task {
	result := make(map[string]models.Task)
	for taskId, task := range core.WorkConfig.Tasks {
		log.Info("current task id: ", taskId)
		if task.Stage == stage {
			result[taskId] = task
		}
	}
	return result
}
