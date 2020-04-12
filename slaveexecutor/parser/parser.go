package parser

import (
	"io/ioutil"

	"github.com/kubitre/diplom/slaveexecutor/models"
	"gopkg.in/yaml.v2"
)

/*ParseObj - parsing from yaml*/
func ParseObj(data []byte) (*models.WorkConfig, error) {
	run := models.WorkConfig{}
	err := yaml.Unmarshal(data, &run)
	if err != nil {
		return nil, err
	}
	return &run, nil
}

/*TestLoadFile - test loading file with test yaml task*/
func TestLoadFile(filepath string) ([]byte, error) {
	file, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	return file, nil
}
