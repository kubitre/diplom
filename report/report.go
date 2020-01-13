package report

type ReportCombiner struct {
	LogsPerStage  map[string][]string
	TimeExecuting map[string][]int64
}

func NewReportCombiner() *ReportCombiner {
	return &ReportCombiner{
		LogsPerStage:  make(map[string][]string),
		TimeExecuting: make(map[string][]int64),
	}
}

func (report *ReportCombiner) AddNewLog(stageName string, logs []string) error {
	return nil
}

func (report *ReportCombiner) AddNewMetric(stageName string, metric []int) error {
	return nil
}
