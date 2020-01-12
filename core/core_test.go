package core

import "testing"

import "github.com/kubitre/diplom/gitmod"

import "github.com/kubitre/diplom/docker"

import "github.com/kubitre/diplom/models"

func TestCreateNewRunnerError(t *testing.T) {
	_, err := NewCoreRunner(nil, nil, 10, nil)
	if err == nil {
		t.Error("success create runner but it not possible. " + err.Error())
	}
	t.Log("completed create runner. ")
}

func TestCreateNewRunnerSuccess(t *testing.T) {
	dock, err := docker.NewDockerExecutor()
	if err != nil {
		t.Error("can not create docker executor. " + err.Error())
	}
	if _, err := NewCoreRunner(&gitmod.Git{}, dock, 10, nil); err != nil {
		t.Error("not created runner." + err.Error())
	}
}

func Test_SetupConfigurationPipeline(t *testing.T) {
	dock, err := docker.NewDockerExecutor()
	if err != nil {
		t.Error("can not create docker executor. " + err.Error())
	}
	runner, err := NewCoreRunner(&gitmod.Git{}, dock, 10, nil)
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.SetupConfigurationPipeline(&models.WorkConfig{
		Stages: []string{
			"test",
		},
		Tasks: map[string]models.Task{
			"Test": models.Task{
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
	dock, err := docker.NewDockerExecutor()
	if err != nil {
		t.Error("can not create docker executor. " + err.Error())
	}
	runner, err := NewCoreRunner(&gitmod.Git{}, dock, 10, nil)
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.SetupConfigurationPipeline(&models.WorkConfig{
		Stages: []string{},
		Tasks: map[string]models.Task{
			"Test": models.Task{
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
	dock, err := docker.NewDockerExecutor()
	if err != nil {
		t.Error("can not create docker executor. " + err.Error())
	}
	runner, err := NewCoreRunner(&gitmod.Git{}, dock, 10, nil)
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.SetupConfigurationPipeline(&models.WorkConfig{
		Stages: []string{
			"test",
		},
		Tasks: map[string]models.Task{},
	}); err != nil {
		t.Log(err)
	} else {
		t.Error()
	}
}

func Test_CreatePipelineError(t *testing.T) {
	dock, err := docker.NewDockerExecutor()
	if err != nil {
		t.Error("can not create docker executor. " + err.Error())
	}
	runner, err := NewCoreRunner(&gitmod.Git{}, dock, 10, nil)
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
	dock, err := docker.NewDockerExecutor()
	if err != nil {
		t.Error("can not create docker executor. " + err.Error())
	}
	runner, err := NewCoreRunner(&gitmod.Git{}, dock, 10, nil)
	if err != nil {
		t.Error("not created runner." + err.Error())
	}
	if err := runner.CreatePipeline(&models.WorkConfig{
		Stages: []string{
			"test",
		},
		Tasks: map[string]models.Task{
			"test": models.Task{
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
