package enhancer

/*MergeMetricsToString - смержить все спарсенные метрики в строки, с разделителями в виде запятой*/
func MergeMetricsToString(metrics map[string][]string) map[string]string {
	result := map[string]string{}
	for nameMetric, values := range metrics {
		result[nameMetric] = mergeStrings(values)
	}
	return result
}

func mergeStrings(need []string)string {
	resulst := ""
	for _, value := range need {
		resulst += value + ", "
	}
	return resulst
}
