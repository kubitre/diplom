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
	"github.com/kubitre/diplom/config"
	"github.com/kubitre/diplom/discovery"
	"github.com/kubitre/diplom/docker_runner"
	"github.com/kubitre/diplom/gitmod"
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/payloads"
	log "github.com/sirupsen/logrus"
)

type (
	/*SlaveRunnerCore - ядро для слейва*/
	SlaveRunnerCore struct {
		Git          *gitmod.Git
		Docker       *docker_runner.DockerExecutor
		WorkerPull   []chan models.TaskConfig
		ChannelClose chan string
		SlaveConfig  *config.ConfigurationSlaveRunner
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
	config *config.ConfigurationSlaveRunner,
	configService *config.ServiceConfig,
) (*SlaveRunnerCore, error) {
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
	configService.SetupNewPort(port)
	log.Println("start initialize discovery module")
	discove := discovery.InitializeDiscovery(discovery.SlavePattern, configService)
	if errClientConsul := discove.NewClientForConsule(); errClientConsul != nil {
		log.Println("can not register client in consul: ", errClientConsul)
		os.Exit(1)
	}
	log.Println("completed initilize discovery module")
	discove.RegisterServiceWithConsul([]string{discovery.TagSlave})
	return &SlaveRunnerCore{
		Git:          &gitmod.Git{},
		Docker:       dock,
		ChannelClose: make(chan string, 1),
		SlaveConfig:  config,
		Discovery:    discove,
	}, nil
}

/*UnregisterService - деаутентификация сервиса в консуле*/
func (core *SlaveRunnerCore) UnregisterService() {
	core.Discovery.UnregisterCurrentService()
}

/*RunWorkers - запуск пула воркеров*/
func (core *SlaveRunnerCore) RunWorkers() {
	runParallelExecutors(core.SlaveConfig.AmountPullWorkers, core.ChannelClose, core)
}

func runParallelExecutors(
	amountParallelExecutors int,
	chanelForClosed chan string,
	core *SlaveRunnerCore) []chan models.TaskConfig {
	log.Println("starting all executing workers")
	resultChannels := make([]chan models.TaskConfig, amountParallelExecutors)
	for i := 0; i < amountParallelExecutors; i++ {
		resultChannels[i] = make(chan models.TaskConfig, 1)
		go executor(i, resultChannels[i], chanelForClosed, *core)
	}
	log.Println("completed start all executing workers")
	return resultChannels
}

func executor(executorID int, taskChallenge chan models.TaskConfig, close chan string, core SlaveRunnerCore) {
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
func (core *SlaveRunnerCore) SetupConfigurationPipeline(config *models.TaskConfig) error {
	if len(config.Stages) == 0 {
		return errors.New("runner config should have 1 or mode stages annotation")
	}
	if len(config.Jobs) == 0 {
		return errors.New("runner config should have 1 or more tasks")
	}
	return nil
}

/*CreatePipeline - создание пайплайна на выполнение одной задачи*/
func (core *SlaveRunnerCore) CreatePipeline(taskConfig *models.TaskConfig) error {
	if taskConfig == nil {
		addressMaster, errAddress := core.getAddressMaster()
		if errAddress != nil {
			return errAddress
		}
		sendStatusTask("http://"+addressMaster+"/task/"+taskConfig.TaskID+"/status", taskConfig.TaskID, models.FAILED)
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

func (core *SlaveRunnerCore) faieldTask(taskID string) {
	log.Info("started sending failed for task to master executor")
	addressMaster, errAddress := core.getAddressMaster()
	if errAddress != nil {
		log.Error("can not get address of master executor")
	}
	sendStatusTask("http://"+addressMaster+"/task/"+taskID+"/status", taskID, models.FAILED)
}

/*executingJobsInStage - sxecute entry for jobs start*/
func (core *SlaveRunnerCore) executingJobsInStage(stage string, taskConfig *models.TaskConfig) []chan int {
	log.Info("start executing jobs in stage: ", stage)
	currentJobs := core.getJobsByStage(stage, taskConfig.Jobs, taskConfig.TaskID)
	if len(currentJobs) == 0 {
		core.faieldTask(taskConfig.TaskID)
		return nil // TODO: ADD error
	}
	jobsChecked := make([]chan int, len(currentJobs))
	for indexJob, job := range currentJobs {
		jobsChecked[indexJob] = make(chan int, 1)
		log.Info("current tasks for stage: ", stage, "; job: ", job.Reports)
		jobResult := make(chan models.LogsPerTask, 1)
		go executingParallelJobPerStage(job, core, jobResult)
		go checkJobResult(jobResult, job, core, jobsChecked[indexJob])
	}
	return jobsChecked
}

func sendStatusTask(address, taskID string, status models.TaskStatusIndx) error {
	log.Info("start sending results to master node")
	pay := payloads.ChangeStatusTask{
		TaskID:    taskID,
		NewStatus: int(status),
	}
	resultMarshal, errMarshal := json.Marshal(&pay)
	if errMarshal != nil {
		return errMarshal
	}
	r := bytes.NewReader(resultMarshal)
	_, err := http.Post(address, "application/json", r)
	if err != nil {
		return err
	}
	return nil
}

func sendStatusJob(address, taskID, jobName string, status models.TaskStatusIndx) error {
	log.Info("start sending job status to master node")
	pay := payloads.ChangeStatusJob{
		TaskID:    taskID,
		NewStatus: int(status),
		Job:       jobName,
	}
	resultMarshal, errMarshal := json.Marshal(&pay)
	if errMarshal != nil {
		return errMarshal
	}
	r := bytes.NewReader(resultMarshal)
	_, err := http.Post(address, "application/json", r)
	if err != nil {
		log.Error("can not sent status to master executor")
		return err
	}
	return nil
}

func (core *SlaveRunnerCore) getAddressMaster() (string, error) {
	allServices := core.Discovery.GetService(discovery.MasterPattern, discovery.TagMaster)
	if len(allServices) == 0 {
		log.Error("not found master executor in consul. Can not sending result")
		return "", errors.New("not found master executor")
	}
	address := allServices[0].Node.Address + ":" + strconv.Itoa(allServices[0].Service.Port)
	return address, nil
}

func checkJobResult(jobResult chan models.LogsPerTask, job models.Job, core *SlaveRunnerCore, jobChecked chan int) {
	log.Info("start checking result work for job: ", job.JobName)
	result := <-jobResult

	allLogs := mergeSTD(result)
	reports := parseSTDToReport(allLogs, job)
	address, errAddress := core.getAddressMaster()
	if errAddress != nil {
		log.Error("not found master executor in consul. Can not sending result")
		jobChecked <- -1
	}
	// add Job output
	if len(result.STDERR) > 0 {
		sendStatusJob("http://"+address+"/task/"+job.TaskID+"/status/"+job.JobName, job.TaskID, job.JobName, models.FAILED)
		sendStatusTask("http://"+address+"/task/"+job.TaskID+"/status", job.TaskID, models.FAILED)
	} else {
		sendStatusJob("http://"+address+"/task/"+job.TaskID+"/status/"+job.JobName, job.TaskID, job.JobName, models.SUCCESS)
	}
	if errSend := sendResultLogsToMaster("http://"+address+"/task/"+job.TaskID+"/log/"+job.Stage+"/"+job.JobName, result); errSend != nil {
		jobChecked <- -1
		return
	}
	if errSend := sendResultReportsToMaster("http://"+address+"/task"+job.TaskID+"/reports/"+job.Stage+"/"+job.JobName, reports); errSend != nil {
		jobChecked <- -1
		return
	}
	jobChecked <- 1
}

func sendResultReportsToMaster(address string, result map[string][]string) error {
	resultMarshal, errMarshal := json.Marshal(&result)
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

/*mergeSTD - merge output from containers in one. Need for create report*/
func mergeSTD(jobResult models.LogsPerTask) (result string) {
	for _, value := range jobResult.STDOUT {
		result += value + "\n"
	}
	for _, value := range jobResult.STDERR {
		result += value + "\n"
	}
	return
}

func parseSTDToReport(allLogs string, job models.Job) map[string][]string {
	result := map[string][]string{} // result in format: key: []values
	log.Println("start extracting data from logs to report by regexp: ", job)
	for nameRegular, rex := range job.Reports {
		log.Println("start extracting report: ", rex, " name: ", nameRegular)
		result[nameRegular] = parseSTD(rex, allLogs)
	}
	return result
}

func parseSTD(regx string, logs string) []string {
	reg := regexp.MustCompile(regx)
	founded := reg.FindStringSubmatch(logs)
	log.Println("founded: ", founded)
	log.Println("sub groups: ", reg.SubexpNames())
	return founded
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

func executingParallelJobPerStage(job models.Job, core *SlaveRunnerCore, jobResult chan models.LogsPerTask) {
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

// DEPRECATED
func (core *SlaveRunnerCore) getRepoCandidate(job models.Job) (string, error) {
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

func (core *SlaveRunnerCore) prepareTask(job models.Job) error {
	// pathRepo, err := core.getRepoCandidate(job)
	// if err != nil {
	// 	return err
	// }
	log.Println("creating image for job: ", job.JobName)
	// log.Println("path repo: ", pathRepo)
	log.Println("name of docker image: ", job.TaskID+"_"+job.JobName)
	if err := core.Docker.CreateImageMem(job.Image,
		job.ShellCommands,
		[]string{strings.ToLower(job.TaskID + "_" + job.JobName)},
		map[string]string{}); err != nil {
		return err
	}
	// if err := os.RemoveAll(pathRepo); err != nil {
	// 	log.Warn("can not remove repo candidate. ", err)
	// }
	return nil
}

// DEPRECATED
func (core *SlaveRunnerCore) appendRepoIntoDocker(path string, job models.Job) []string {
	return append(job.Image, `COPY `+path+` /repoCandidate`)
}

/*getJobsByStage - получение всех исполняемых job на stage */
func (core *SlaveRunnerCore) getJobsByStage(stage string, jobs map[string]models.Job, taskID string) []models.Job {
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
			Reports:             job.Reports,
		}
		if job.Stage == stage {
			result = append(result, enhanceJob)
		}
	}
	log.Println("jobs per stage: ", result)
	return result
}
