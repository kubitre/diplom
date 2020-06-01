package tools

/*AppendMap - добавить к карте карту*/
func AppendMap(current map[string]string, need map[string]string) map[string]string {
	resultMap := current
	for nameField, result := range need {
		settingUp := false
		for nameCurrent := range resultMap {
			if nameField == nameCurrent {
				settingUp = true
				break
			}
		}
		if !settingUp {
			resultMap[nameField] = result
		}
	}
	return resultMap
}
