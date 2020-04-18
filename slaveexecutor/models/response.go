package models

type (
	// LogsPerTask - логи по выполнению какой-либо джобы
	LogsPerTask struct {
		STDOUT []string
		STDERR []string
	}

	// ReportPerTask - модель отчёта для мастера на уровне джобы
	ReportPerTask struct {
		Result map[string][]string
	}
)
