package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/kubitre/diplom/slaveexecutor/config"
	"github.com/kubitre/diplom/slaveexecutor/discovery"
	"github.com/kubitre/diplom/slaveexecutor/docker_runner"
	"github.com/kubitre/diplom/slaveexecutor/gitmod"
	"github.com/kubitre/diplom/slaveexecutor/models"
	log "github.com/sirupsen/logrus"
)

type (
	/*CoreSlaveRunner - ядро для слейва*/
	CoreSlaveRunner struct {
		Git          *gitmod.Git
		Docker       *docker_runner.DockerExecutor
		WorkerPull   []chan models.TaskConfig
		ChannelClose chan string
		SlaveConfig  *config.SlaveConfiguration
		Discovery    *discovery.Discovery
	}
	/*Worker - единичная воркер функция, которая отвечает за выполнение всех job на одной стадии одной задачи*/
	Worker struct {
		Task   chan models.Job         // текущая задача у воркера
		Result chan models.LogsPerTask // текущий результат у воркера (лог файл)
	}
)

/*NewCoreSlaveRunner - инициализация нового ядра слейв модуля*/
func NewCoreSlaveRunner(
	config *config.SlaveConfiguration,
) (*CoreSlaveRunner, error) {
	dock, err := docker_runner.NewDockerExecutor()
	if err != nil {
		log.Error("can not create docker executor. " + err.Error())
	}
	log.Println("initialize new port")
	port, errPort := discovery.GetAvailablePort()
	if errPort != nil {
		log.Println("can not initialize new port: ", errPort)
		os.Exit(1)
	}
	config.SetupNewPort(port)
	log.Println("start initialize discovery module")
	discove := discovery.InitializeDiscovery(config)
	if errClientConsul := discove.NewClientForConsule(); errClientConsul != nil {
		log.Println("can not register client in consul: ", errClientConsul)
		os.Exit(1)
	}
	log.Println("completed initilize discovery module")
	discove.RegisterServiceWithConsul()
	return &CoreSlaveRunner{
		Git:          &gitmod.Git{},
		Docker:       dock,
		ChannelClose: make(chan string, 1),
		SlaveConfig:  config,
		Discovery:    discove,
	}, nil
}

/*RunWorkers - запуск пула воркеров*/
func (core *CoreSlaveRunner) RunWorkers() {
	runParallelExecutors(core.SlaveConfig.AmountPullWorkers, core.ChannelClose, core)
}

func runParallelExecutors(
	amountParallelExecutors int,
	chanelForClosed chan string,
	core *CoreSlaveRunner) []chan models.TaskConfig {
	log.Println("starting all executing workers")
	resultChannels := make([]chan models.TaskConfig, amountParallelExecutors)
	for i := 0; i < amountParallelExecutors; i++ {
		resultChannels[i] = make(chan models.TaskConfig, 1)
		go executor(i, resultChannels[i], chanelForClosed, *core)
	}
	log.Println("completed start all executing workers")
	return resultChannels
}

func executor(executorID int, taskChallenge chan models.TaskConfig, close chan string, core CoreSlaveRunner) {
	for {
		select {
		case close := <-close:
			log.Info("stop worker: ", executorID, " by closed signal: ", close)
		case newTask := <-taskChallenge:
			log.Debug("start working with new task: ", newTask)
			if err := core.CreatePipeline(&newTask); err != nil {
				log.Error("can not create pipeline for task. Err: ", err)
			}
			// send to Master node result log
		}
	}
}

/*SetupConfigurationPipeline - setting up configuration if its no configuring in start
NOW USING ONLY FOR TESTING
*/
func (core *CoreSlaveRunner) SetupConfigurationPipeline(config *models.TaskConfig) error {
	if len(config.Stages) == 0 {
		return errors.New("runner config should have 1 or mode stages annotation")
	}
	if len(config.Jobs) == 0 {
		return errors.New("runner config should have 1 or more tasks")
	}
	return nil
}

/*CreatePipeline - создание пайплайна на выполнение одной задачи*/
func (core *CoreSlaveRunner) CreatePipeline(taskConfig *models.TaskConfig) error {
	if taskConfig == nil {
		return errors.New("can not create pipeline without configuration. Please setup configuration and continue")
	}
	for _, stage := range taskConfig.Stages {
		log.Info("start working on stage: " + stage)
		checked := core.executingJobsInStage(stage, taskConfig)
		for _, check := range checked {
			<-check
		}
	}
	return nil
}

func (core *CoreSlaveRunner) executingJobsInStage(stage string, taskConfig *models.TaskConfig) []chan int {
	log.Println("start executing jobs in stage: ", stage)
	currentJobs := core.getJobsByStage(stage, taskConfig.Jobs, taskConfig.TaskID)
	jobsChecked := make([]chan int, len(currentJobs))
	for indexJob, job := range currentJobs {
		jobsChecked[indexJob] = make(chan int, 1)
		log.Info("current tasks for stage: ", stage, "; job: ", job)
		jobResult := make(chan models.LogsPerTask, 1)
		go executingParallelJobPerStage(job, core, jobResult)
		go checkJobResult(jobResult, job, core, jobsChecked[indexJob])
	}
	return jobsChecked
}

func checkJobResult(jobResult chan models.LogsPerTask, job models.Job, core *CoreSlaveRunner, jobChecked chan int) {
	log.Println("start checking result work for job: ", job.JobName)
	result := <-jobResult
	allLogs := mergeSTD(result)
	parseSTDToReport(allLogs, job)
	allServices := core.Discovery.GetService("master-executor", "master")
	address := allServices[0].Address + ":" + strconv.Itoa(allServices[0].ServicePort)
	// add Job output
	if errSend := sendResultLogsToMaster("http://"+address+"/task/"+job.TaskID+"/log/"+job.Stage+"/"+job.JobName, result); errSend != nil {
		jobChecked <- -1
		return
	}
	jobChecked <- 1
}

func mergeSTD(jobResult models.LogsPerTask) (result string) {
	for _, value := range jobResult.STDOUT {
		result += value + "\n"
	}
	for _, value := range jobResult.STDERR {
		result += value + "\n"
	}
	return
}

func parseSTDToReport(allLogs string, job models.Job) {
	log.Println("start extracting data from logs to report by regexp: ", job)
	for nameRegular, rex := range job.Reports {
		log.Println("start extracting report: ", rex, " name: ", nameRegular)
		parseSTD(rex, allLogs)
	}
}

func parseSTD(regx string, logs string) {
	reg := regexp.MustCompile(regx)
	founded := reg.FindStringSubmatch(logs)
	log.Println("founded: ", founded)
	log.Println("sub groups: ", reg.SubexpNames)
}

func sendResultLogsToMaster(address string, jobResult models.LogsPerTask) error {
	resultMarshal, errMarshal := json.Marshal(&jobResult)
	if errMarshal != nil {
		log.Println("can not marshaled response: ", errMarshal)
		return errMarshal
	}
	r := bytes.NewReader(resultMarshal)
	resp, err := http.Post(address, "application/json", r)
	if err != nil {
		log.Error("can not execute request", err)
		return err
	}
	log.Debug("response from master: ", resp)
	return nil
}

func readSTD(buffer *bytes.Buffer) []string {
	var result []string
	for {
		line, err := buffer.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
		}
		result = append(result, line)
	}
	return result
}

func executingParallelJobPerStage(job models.Job, core *CoreSlaveRunner, jobResult chan models.LogsPerTask) {
	log.Println("start preparing job: ", job.JobName)
	if err := core.prepareTask(job); err != nil {
		log.Error("error while preparing task. ", err)
	}
	log.Println("start creating container for job: ", job.JobName)
	containername := strings.ToLower(job.TaskID + "_" + job.JobName)
	core.Docker.RemoveContainer("execute_" + containername)
	containerID, err := core.Docker.CreateContainer(&models.ContainerCreatePayload{
		BaseImageName: containername,
		ContainerName: "execute_" + containername,
	})
	log.Println("running container for job")
	responseCloser, err := core.Docker.RunContainer(containerID)
	log.Println("error while starting container: ", err)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	_, err = stdcopy.StdCopy(stdout, stderr, responseCloser)
	// ADD PARSING STDOUT and STDERR
	output := models.LogsPerTask{
		STDERR: readSTD(stderr),
		STDOUT: readSTD(stdout),
	}
	log.Println("output: ", output)
	defer responseCloser.Close()
	if err != nil {
		jobResult <- models.LogsPerTask{}
	}
	jobResult <- output
}

func (core *CoreSlaveRunner) getRepoCandidate(job models.Job) (string, error) {
	log.Println("start cloning repository candidate")
	path, err := core.Git.CloneRepo(job.RepositoryCandidate)
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

func (core *CoreSlaveRunner) prepareTask(job models.Job) error {
	pathRepo, err := core.getRepoCandidate(job)
	if err != nil {
		return err
	}
	log.Println("creating image for job: ", job.JobName)
	log.Println("path repo: ", pathRepo)
	log.Println("name of docker image: ", job.TaskID+"_"+job.JobName)
	if err := core.Docker.CreateImageMem(core.appendRepoIntoDocker(pathRepo, job),
		job.ShellCommands,
		[]string{strings.ToLower(job.TaskID + "_" + job.JobName)},
		map[string]string{}); err != nil {
		return err
	}
	if err := os.RemoveAll(pathRepo); err != nil {
		log.Warn("can not remove repo candidate. ", err)
	}
	return nil
}

func (core *CoreSlaveRunner) appendRepoIntoDocker(path string, job models.Job) []string {
	return append(job.Image, `COPY `+path+` /repoCandidate`)
}

/*getJobsByStage - получение всех исполняемых job на stage */
func (core *CoreSlaveRunner) getJobsByStage(stage string, jobs map[string]models.Job, taskID string) []models.Job {
	result := []models.Job{}
	log.Println("getting all jobs for stage: ", stage)
	for jobID, job := range jobs {
		log.Info("current task id: ", jobID)
		enhanceJob := models.Job{
			JobName:             jobID,
			Stage:               stage,
			TaskID:              taskID,
			Image:               job.Image,
			RepositoryCandidate: job.RepositoryCandidate,
			ShellCommands:       job.ShellCommands,
		}
		if job.Stage == stage {
			result = append(result, enhanceJob)
		}
	}
	log.Println("jobs per stage: ", result)
	return result
}
