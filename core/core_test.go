package core

import (
	"testing"

	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/gitmod"
	"github.com/kubitre/diplom/models"
)

func TestCreateNewRunnerSuccess(t *testing.T) {
	if _, err := NewCoreSlaveRunner(&config.ConfigurationSlaveRunner{
		AmountPullWorkers:          10,
		AmountParallelTaskPerStage: 100,
	}, &config.ServiceConfig{}); err != nil {
		t.Log("not created runner." + err.Error())
	}
}

func Test_SetupConfigurationPipeline(t *testing.T) {

	runner, err := NewCoreSlaveRunner(
		&config.ConfigurationSlaveRunner{
			AmountPullWorkers:          10,
			AmountParallelTaskPerStage: 100,
		}, &config.ServiceConfig{},
	)
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.SetupConfigurationPipeline(&models.TaskConfig{
		Stages: []string{
			"test",
		},
		Jobs: map[string]models.Job{
			"Test": models.Job{
				Stage: "test",
				ShellCommands: []string{
					"ls -la",
				},
			},
		},
	}); err != nil {
		t.Error(err)
	}
}

func Test_SetupConfigurationPipelineZeroStages(t *testing.T) {

	runner, err := NewCoreSlaveRunner(&config.ConfigurationSlaveRunner{
		AmountPullWorkers:          10,
		AmountParallelTaskPerStage: 100,
	}, &config.ServiceConfig{})
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.SetupConfigurationPipeline(&models.TaskConfig{
		Stages: []string{},
		Jobs: map[string]models.Job{
			"Test": models.Job{
				Stage: "test",
				ShellCommands: []string{
					"ls -la",
				},
			},
		},
	}); err != nil {
		t.Log(err)
	} else {
		t.Error()
	}
}

func Test_SetupConfigurationPipelineZeroTasks(t *testing.T) {

	runner, err := NewCoreSlaveRunner(&config.ConfigurationSlaveRunner{
		AmountPullWorkers:          10,
		AmountParallelTaskPerStage: 100,
	}, &config.ServiceConfig{})
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.SetupConfigurationPipeline(&models.TaskConfig{
		Stages: []string{
			"test",
		},
		Jobs: map[string]models.Job{},
	}); err != nil {
		t.Log(err)
	} else {
		t.Error()
	}
}

func Test_CreatePipelineError(t *testing.T) {

	runner, err := NewCoreSlaveRunner(&config.ConfigurationSlaveRunner{
		AmountPullWorkers:          10,
		AmountParallelTaskPerStage: 100,
	}, &config.ServiceConfig{})
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.CreatePipeline(nil); err != nil {
		t.Log("completed test.", err)
	} else {
		t.Error("")
	}
}

func Test_CreatePipelineWithConfig(t *testing.T) {

	runner, err := NewCoreSlaveRunner(&config.ConfigurationSlaveRunner{
		AmountPullWorkers:          10,
		AmountParallelTaskPerStage: 100,
	}, &config.ServiceConfig{})
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.CreatePipeline(&models.TaskConfig{
		Stages: []string{
			"test",
		},
		Jobs: map[string]models.Job{
			"test": models.Job{
				Stage:               "test",
				RepositoryCandidate: "https://github.com/kubitre/for_diplom",
				ShellCommands: []string{
					"ls -la",
					"cat test.txt",
				},
				Image: []string{
					"FROM ubuntu:18.04 as runnerContext",
					`RUN echo "test" > test.txt`,
					"RUN cat test.txt",
				},
			},
		},
	}); err != nil {
		t.Error(err)
	}
	t.Log("completed test.")
}

func Test_GetTypeError(t *testing.T) {
	git := gitmod.Git{}
	path, err := git.CloneRepo("http://github.com/kubitre/")
	switch git.GetTypeError(err) {
	case gitmod.ErrorAuthenticate:
		t.Log("test")
		t.Error()
		return
	case gitmod.ErrorExistingRepository:
		t.Log("test")
		t.Error()
		return
	case gitmod.ErrorUnrecognized:
		t.Log("test")
		return
	}
	t.Error("Path: ", path)
}
