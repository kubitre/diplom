package core

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
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
		WorkerPull   chan models.TaskConfig
		ChannelClose chan string
		SlaveConfig  *config.ConfigurationSlaveRunner
		Discovery    *discovery.Discovery
	}
	/*Worker - единичная воркер функция, которая отвечает за выполнение всех job на одной стадии одной задачи*/
	Worker struct {
		Task   chan models.Job         // текущая задача у воркера
		Result chan models.LogsPerTask // текущий результат у воркера (лог файл)
	}
	/*WorkJob - статус по задаче*/
	WorkJob struct {
		JobName    string
		Stage      string
		TaskID     string
		JobStatus  int
		JobResukt  models.LogsPerTask
		JobReports models.ReportPerTask
		JobMetrics map[string]string
	}
)

const (
	failJob     = 0
	executedJob = 1
	successJob  = 2
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
	core.WorkerPull = runParallelExecutors(core.SlaveConfig.AmountPullWorkers, core.ChannelClose, core)
}

func runParallelExecutors(
	amountParallelExecutors int,
	chanelForClosed chan string,
	core *SlaveRunnerCore) chan models.TaskConfig {
	log.Println("starting all executing workers")
	tasksPool := make(chan models.TaskConfig, amountParallelExecutors)
	for i := 0; i < amountParallelExecutors; i++ {
		go executor(i, tasksPool, chanelForClosed, *core)
	}
	log.Println("completed start all executing workers")
	return tasksPool
}

func executor(executorID int, taskChallenge <-chan models.TaskConfig, close chan string, core SlaveRunnerCore) {
	for {
		select {
		case close := <-close:
			log.Info("stop worker: ", executorID, " by closed signal: ", close)
		case newTask := <-taskChallenge:
			log.Debug("start working with new task: ", newTask, " on worker : ", executorID)
			if err := core.CreatePipeline(&newTask); err != nil {
				log.Error("can not create pipeline for task. Err: ", err)
				core.faieldTask(newTask.TaskID, "unknown")
			} else {
				core.successTask(newTask.TaskID, "unknown")
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
		return errors.New("can not create pipeline without configuration. Please setup configuration and continue")
	}
	log.Debug("All available stages: ", taskConfig.Stages)
	for _, stage := range taskConfig.Stages {
		log.Info("start working on stage: " + stage)
		jobWork, amountJobs, err := core.executingJobsInStage(stage, taskConfig)
		if err != nil {
			log.Error("Something went wrong while exeucing jobs in stage: ", stage)
			return err
		}
		for i := 0; i < amountJobs; i++ {
			result := <-jobWork
			if errChecking := checkJobResult(result, core); errChecking != nil {
				return errChecking
			}
		}
	}
	return nil
}

func (core *SlaveRunnerCore) faieldTask(taskID, stage string) {
	log.Debug("start send statu Failed for task to Master")
	core.sendStatusTaskToMaster(taskID, models.FAILED, stage)
}

func (core *SlaveRunnerCore) startTask(taskID, stage string) {
	log.Debug("start send status Running task to Master")
	core.sendStatusTaskToMaster(taskID, models.RUNNING, stage)
}

func (core *SlaveRunnerCore) successTask(taskID, stage string) {
	log.Debug("start send status Success task to Master")
	core.sendStatusTaskToMaster(taskID, models.SUCCESS, stage)
}

func (core *SlaveRunnerCore) failedJob(taskID string, jobName string) {
	log.Debug("start send status Fail for job to Master")
	core.sendStatusJobToMaster(taskID, jobName, models.FAILED)
}
func (core *SlaveRunnerCore) successJob(taskID, jobName string) {
	log.Debug("start send status Success for job to Master")
	core.sendStatusJobToMaster(taskID, jobName, models.SUCCESS)
}

func (core *SlaveRunnerCore) sendStatusTaskToMaster(taskID string, status models.TaskStatusIndx, stage string) {
	log.Info("started sending status for task to master executor")
	addressMaster, errAddress := core.getAddressMaster()
	if errAddress != nil {
		log.Error("can not get address of master executor")
	}
	if errStatusTask := sendStatusTask("http://"+addressMaster+"/task/"+taskID+"/status", taskID, status, stage); errStatusTask != nil {
		log.Error("Can not send status task: ", errStatusTask)
	}
}

func (core *SlaveRunnerCore) sendStatusJobToMaster(taskID, jobName string, status models.TaskStatusIndx) {
	log.Info("started sending status for job to master executor")
	addressMaster, errAddress := core.getAddressMaster()
	if errAddress != nil {
		log.Error("can not get address of master executor")
	}
	if errStatusJob := sendStatusJob("http://"+addressMaster, taskID, jobName, status); errStatusJob != nil {
		log.Error("Can not send status job: ", errStatusJob)
	}
}

/*executingJobsInStage - sxecute entry for jobs start*/
func (core *SlaveRunnerCore) executingJobsInStage(stage string, taskConfig *models.TaskConfig) (chan WorkJob, int, error) {
	log.Info("start executing jobs in stage: ", stage)
	log.Debug("ALL JOBS: ", taskConfig.Jobs)
	currentJobs := core.getJobsByStage(stage, taskConfig.Jobs, taskConfig.TaskID)
	if len(currentJobs) == 0 {
		core.faieldTask(taskConfig.TaskID, "unknown")
		return nil, 0, errors.New("can not executing task, because jobs was empty") // TODO: ADD error
	}
	core.startTask(taskConfig.TaskID, stage)
	jobWork := make(chan WorkJob, len(currentJobs))
	// jobsChecked := make(chan int, len(currentJobs))
	// jobsResult := make(chan models.LogsPerTask, len(currentJobs))
	// jobsNames := make(chan string, len(currentJobs))
	for _, job := range currentJobs {
		log.Info("current tasks for stage: ", stage, "; job: ", job.Reports)
		go executingParallelJobPerStage(job, core, jobWork)
		// go checkJobResult(jobWork, job, core, jobsChecked, jobsNames)
	}
	return jobWork, len(currentJobs), nil
}

func sendStatusTask(address, taskID string, status models.TaskStatusIndx, stage string) error {
	log.Info("start sending results to master node")
	pay := payloads.ChangeStatusTask{
		TaskID:       taskID,
		NewStatus:    int(status),
		CurrentStage: stage,
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
		log.Error("can not sent status to master executor: ", err, " Params: ", address, taskID, jobName, status)
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

func (core *SlaveRunnerCore) extractLogs(workJob WorkJob) error {
	if errMetricExtract := core.extractMetrtics(workJob); errMetricExtract != nil {
		return errMetricExtract
	}
	address, errAddress := core.getAddressMaster()
	if errAddress != nil {
		log.Error("not found master executor in consul. Can not sending result")
	}
	if len(workJob.JobResukt.STDERR) > 0 {
		log.Debug("job : ", workJob.JobName, " was failed status, because have stderrs")
		sendStatusJob("http://"+address+"/task/"+workJob.TaskID+"/status/"+workJob.JobName, workJob.TaskID, workJob.JobName, models.FAILED)
		core.faieldTask(workJob.TaskID, workJob.Stage)
		return errors.New("can not send result to master executor")
	}
	sendStatusJob("http://"+address+"/task/"+workJob.TaskID+"/status/"+workJob.JobName, workJob.TaskID, workJob.JobName, models.SUCCESS)
	if errSend := sendResultLogsToMaster("http://"+address+"/task/"+workJob.TaskID+"/log/"+workJob.Stage+"/"+workJob.JobName, workJob.JobResukt); errSend != nil {
		log.Error("can not sending result to master: ", errSend)
		return errSend
	}
	return nil
}

func (core *SlaveRunnerCore) extractMetrtics(workJob WorkJob) error {
	log.Debug("start extracting metics from logs")
	allLogs := mergeSTD(workJob.JobResukt)
	log.Debug("all logs: ", allLogs, " reg: ", workJob.JobMetrics)
	reports := parseSTDToReport(allLogs, workJob.JobMetrics)
	log.Debug("parsed metrics: ", reports)
	address, errAddress := core.getAddressMaster()
	if errAddress != nil {
		log.Error("not found master executor in consul. Can not sending result")
	}
	if errSend := sendResultReportsToMaster("http://"+address+"/task/"+workJob.TaskID+"/reports/"+workJob.JobName, reports); errSend != nil {
		log.Error("can not send report to master executor")
		// jobChecked <- failJob
		// jobName <- job.JobName
		return errSend
	}
	return nil
}

func checkJobResult(jobWork WorkJob, core *SlaveRunnerCore) error {
	log.Info("start checking result work for job: ", jobWork.JobName)
	switch jobWork.JobStatus {
	case failJob:
		log.Error("error while executing job. start failing task")
		core.failedJob(jobWork.TaskID, jobWork.JobName)
		core.faieldTask(jobWork.TaskID, jobWork.Stage)
		return errors.New("error while executing job. start failing task")
	case executedJob:
		log.Debug("success executing job. sending report per job to master")
		core.successJob(jobWork.TaskID, jobWork.JobName)
		if errExtract1 := core.extractLogs(jobWork); errExtract1 != nil {
			return errExtract1
		}
		if errExtract2 := core.extractMetrtics(jobWork); errExtract2 != nil {
			return errExtract2
		}
		return nil
	default:
		log.Error("Can not recognize status job. Send status failed")
		core.faieldTask(jobWork.TaskID, jobWork.Stage)
		return errors.New("something went wrong, while executing task. Stop executing task with id: " + jobWork.TaskID)
	}
}

func sendResultReportsToMaster(address string, result map[string][]string) error {
	resultMarshal, errMarshal := json.Marshal(&result)
	if errMarshal != nil {
		log.Error("can not marshaled response: ", errMarshal)
		return errMarshal
	}
	r := bytes.NewReader(resultMarshal)
	resp, err := http.Post(address, "application/json", r)
	if err != nil {
		log.Error("can not execute request", err)
		return err
	}
	body, _ := ioutil.ReadAll(resp.Body)
	log.Debug("response from master: ", string(body))
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

func parseSTDToReport(allLogs string, jobsMetrics map[string]string) map[string][]string {
	result := map[string][]string{} // result in format: key: []values
	for nameRegular, rex := range jobsMetrics {
		log.Debug("start extracting report: ", rex, " name: ", nameRegular)
		parsingValues := parseSTD(rex, allLogs)
		log.Debug("parsed metrics: ", parsingValues)
		result[nameRegular] = parsingValues
	}
	return result
}

func parseSTD(regx string, logs string) []string {
	reg := regexp.MustCompile(regx)
	log.Debug("start parsing metrics: regx: ", regx, " logs:", logs)
	founded := reg.FindStringSubmatch(logs)
	log.Debug("founded: ", founded)
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

func executingParallelJobPerStage(job models.Job, core *SlaveRunnerCore, workJob chan WorkJob) {
	log.Debug("start preparing job: ", job.JobName)
	logsFromBuild, imageName, err := core.prepareTask(job)
	defer core.removeImage(imageName)
	if err != nil {
		log.Error("error while preparing task. ", err)
		workJob <- WorkJob{
			JobName:   job.JobName,
			JobStatus: failJob,
			Stage:     job.Stage,
			TaskID:    job.TaskID,
			JobResukt: models.LogsPerTask{
				STDERR: []string{
					err.Error(),
				},
			},
			JobMetrics: job.Reports,
		}
		return
	}
	log.Debug("start creating container for job: ", job.JobName)
	containername := strings.ToLower(job.TaskID + "_" + job.JobName)
	core.Docker.RemoveContainer("execute_" + containername)
	containerID, err := core.Docker.CreateContainer(&models.ContainerCreatePayload{
		BaseImageName: containername,
		ContainerName: "execute_" + containername,
	})
	if err != nil {
		log.Error("can not create container: ", err)
		workJob <- WorkJob{
			JobName:   job.JobName,
			JobStatus: failJob,
			Stage:     job.Stage,
			TaskID:    job.TaskID,
			JobResukt: models.LogsPerTask{
				STDERR: []string{
					err.Error(),
				},
			},
			JobMetrics: job.Reports,
		}
		return
	}
	log.Debug("running container for job")
	responseCloser, err := core.Docker.RunContainer(containerID, job.Timeout)
	if err != nil {
		log.Error("can not run container: ", err)
		workJob <- WorkJob{
			JobName:   job.JobName,
			JobStatus: failJob,
			Stage:     job.Stage,
			TaskID:    job.TaskID,
			JobResukt: models.LogsPerTask{
				STDERR: []string{
					err.Error(),
				},
			},
			JobMetrics: job.Reports,
		}
		return
	}
	// log.Println("error while starting container: ", err)
	stdout := new(bytes.Buffer)
	stderr := new(bytes.Buffer)
	_, err = stdcopy.StdCopy(stdout, stderr, responseCloser)
	// ADD PARSING STDOUT and STDERR
	resultFromSTD := readSTD(stdout)
	resultFromSTD = append(logsFromBuild, resultFromSTD...)
	output := models.LogsPerTask{
		STDERR: readSTD(stderr),
		STDOUT: resultFromSTD,
	}
	// log.Println("output: ", output)
	defer responseCloser.Close()
	if err != nil {
		workJob <- WorkJob{
			JobName:   job.JobName,
			JobStatus: failJob,
			Stage:     job.Stage,
			TaskID:    job.TaskID,
			JobResukt: models.LogsPerTask{
				STDERR: []string{
					err.Error(),
				},
			},
			JobMetrics: job.Reports,
		}
		return
	}
	workJob <- WorkJob{
		JobName:    job.JobName,
		JobStatus:  executedJob,
		Stage:      job.Stage,
		TaskID:     job.TaskID,
		JobResukt:  output,
		JobMetrics: job.Reports,
	}
}

func (core *SlaveRunnerCore) removeImage(imageName string) {
	log.Debug("REMOVE IMAGE: ", imageName)
	core.Docker.RemoveImage(imageName)
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

func (core *SlaveRunnerCore) prepareTask(job models.Job) ([]string, string, error) {
	// pathRepo, err := core.getRepoCandidate(job)
	// if err != nil {
	// 	return err
	// }
	log.Debug("creating image for job: ", job.JobName)
	// log.Println("path repo: ", pathRepo)
	// log.Println("name of docker image: ", job.TaskID+"_"+job.JobName)
	logsFromBuildStage, err := core.Docker.CreateImageMem(job.Image,
		job.ShellCommands,
		[]string{strings.ToLower(job.TaskID + "_" + job.JobName)},
		map[string]string{})
	if err != nil {
		return []string{}, "", err
	}
	// if err := os.RemoveAll(pathRepo); err != nil {
	// 	log.Warn("can not remove repo candidate. ", err)
	// }
	return logsFromBuildStage, strings.ToLower(job.TaskID + "_" + job.JobName), nil
}

// DEPRECATED
func (core *SlaveRunnerCore) appendRepoIntoDocker(path string, job models.Job) []string {
	return append(job.Image, `COPY `+path+` /repoCandidate`)
}

/*getJobsByStage - получение всех исполняемых job на stage */
func (core *SlaveRunnerCore) getJobsByStage(stage string, jobs map[string]models.Job, taskID string) []models.Job {
	result := []models.Job{}
	log.Debug("getting all jobs for stage: ", stage)
	for jobID, job := range jobs {
		log.Debug("current job id: ", jobID)
		enhanceJob := models.Job{
			JobName:             jobID,
			Stage:               stage,
			TaskID:              taskID,
			Image:               job.Image,
			RepositoryCandidate: job.RepositoryCandidate,
			ShellCommands:       job.ShellCommands,
			Reports:             job.Reports,
			Timeout:             job.Timeout,
		}
		log.Debug("stage: ", stage, " job stage: ", job.Stage)
		if job.Stage == stage {
			result = append(result, enhanceJob)
		}
	}
	log.Debug("jobs per stage: ", result)
	return result
}
