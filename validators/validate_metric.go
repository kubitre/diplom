package validators

import (
	"errors"
)

/*ValidateMetricsForPortal - валидация метрик к типу портала*/
func ValidateMetricsForPortal(metrics map[string][]string) (map[string]string, error) {
	result := map[string]string{}
	for nameMetric, values := range metrics {
		if len(values) > 1 {
			return nil, errors.New("parsed metrics contain more that one metric")
		}
		if len(values) == 1 {
			result[nameMetric] = values[0]
		}
	}
	return result, nil
}
