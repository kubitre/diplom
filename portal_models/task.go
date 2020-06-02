package portal_models

import (
	"strings"

	"github.com/kubitre/diplom/models"
)

type (
	/*PortalTask - формальное описание задачи, приходящей из портала*/
	PortalTask struct {
		TaskID    string     `json:"id"`
		JobGroups []JobGroup `json:"job_groups"`
	}

	/*JobGroup - группа джоб*/
	JobGroup struct {
		NameGroup string `json:"name"`
		Order     int    `json:"order"`
		Jobs      []Job  `json:"jobs"`
	}

	/*Job - джоба*/
	Job struct {
		JobName    string   `json:"name"`
		Dockerfile string   `json:"docker_file"`
		Timeout    int64    `json:"timeout"`
		Metrics    []Metric `json:"metrics"`
	}

	/*Metric - метрики для отчёта*/
	Metric struct {
		MetricName string `json:"key"`
		Regex      string `json:"regex"`
	}
)

// ConvertToAgentTask - конвертер в модель агента
func (task *PortalTask) ConvertToAgentTask() models.TaskConfig {
	needModel := models.TaskConfig{
		TaskID: task.TaskID,
	}
	stages := []string{}
	jobs := map[string]models.Job{}
	groupsJob := task.JobGroups
	sortGroupsByOrder(groupsJob)
	for _, group := range groupsJob {
		stages = append(stages, group.NameGroup)
		jobs = task.appendMap(jobs, task.convertJobs(group.Jobs, task.TaskID, group.NameGroup))
	}
	needModel.Jobs = jobs
	needModel.Stages = stages
	return needModel
}

func (task *PortalTask) appendMap(currentResult map[string]models.Job, newMap map[string]models.Job) map[string]models.Job {
	enhancedResult := currentResult
	for firstName, values := range newMap {
		entered := false
		for secondName := range currentResult {
			if firstName == secondName {
				entered = true
			}
		}
		if !entered {
			enhancedResult[firstName] = values
		}
	}
	return enhancedResult
}

/*convertJobs - конвертация job в тип исполняющего модуля*/
func (task *PortalTask) convertJobs(jobs []Job, taskID, stageName string) map[string]models.Job {
	result := map[string]models.Job{}
	for _, job := range jobs {
		result[job.JobName] = job.convertToAgent(taskID, stageName)
	}
	return result
}

/*convertToAgent - конвертирование конкретной Job в модель исполняющего модуля*/
func (job *Job) convertToAgent(taskID, stageName string) models.Job {
	return models.Job{
		JobName: job.JobName,
		Reports: job.convertMetricsToMap(),
		Image:   job.convertToImage(),
		Stage:   stageName,
		Timeout: job.Timeout,
		TaskID:  taskID,
	}
}

/*convertMetricsToMap - конвертирование необходимых метрик, которые надо парсить из логов*/
func (job *Job) convertMetricsToMap() map[string]string {
	result := map[string]string{}
	for _, metric := range job.Metrics {
		result[metric.MetricName] = metric.Regex
	}
	return result
}

/*convertToImage - конвертирование докерфайла в строковом представлении к массиву строк*/
func (job *Job) convertToImage() []string {
	return strings.Split(job.Dockerfile, "\n")
}

func sortGroupsByOrder(groups []JobGroup) {
	for gap := len(groups) / 2; gap > 0; gap /= 2 {
		for i := gap; i < len(groups); i++ {
			x := groups[i]
			j := i
			for ; j >= gap && groups[j-gap].Order > x.Order; j -= gap {
				groups[j] = groups[j-gap]
			}
			groups[j] = x
		}
	}
}
