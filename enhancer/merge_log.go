package enhancer

import (
	"bufio"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/*Mergelog - merging log in one task path, one stage path
 */
func Mergelog(rootPath, taskID, stage, job string) (string, error) {
	globalLog := rootPath + "/" + taskID + "/"
	if job != "" {
		// add check for stage non empty
		return globalLog + stage + "/" + job + ".log", nil
	}
	if stage != "" {
		return prepareLogs(rootPath + "/" + taskID + "/" + stage)
	}
	if taskID != "" {
		return prepareLogs(rootPath + "/" + taskID)
	}
	return "", errors.New("can not merge log with empty task, stage, job")
}

func walkDir(dirName string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirName, func(path string, info os.FileInfo, err error) error {
		files = append(files, path)
		return nil
	})
	if err != nil {
		log.Println("can not walk by : ", err)
		return nil, err
	}
	return files, nil
}

func prepareLogs(needPathName string) (string, error) {
	walkedFiles, errWalk := walkDir(needPathName)
	if errWalk != nil {
		return "", errWalk
	}
	result := readsFromFiles(walkedFiles)
	resultFileName := needPathName + "_merge.log"
	writeToFile(resultFileName, result)
	return resultFileName, nil
}

func readsFromFiles(files []string) map[string][]string {
	result := map[string][]string{}
	for _, file := range files {
		fileContent, errReading := readFromFile(file)
		if errReading != nil {
			log.Println("can not read file: ", file, " by error: ", errReading)
			continue
		}
		result[file] = fileContent
	}
	return result
}

func validateFile(fileName string) error {
	if !strings.Contains(fileName, ".log") {
		return errors.New("it's no file")
	}
	if strings.Contains(fileName, "_merge") {
		return errors.New("already merged file not using for read")
	}
	return nil
}

func readFromFile(fileName string) ([]string, error) {
	if err := validateFile(fileName); err != nil {
		return nil, err
	}
	file, errOpen := os.Open(fileName)
	if errOpen != nil {
		return nil, errOpen
	}
	defer file.Close()
	result := []string{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		result = append(result, scanner.Text())
	}
	log.Println("reading from file: ", file, " content: ", result)
	return result, nil
}

func writeToFile(filename string, data map[string][]string) {
	file, err := os.Create(filename)
	if err != nil {
		log.Println("can not create file by err: ", err)
		return
	}

	writer := bufio.NewWriter(file)

	for file, contentLine := range data {
		writeWithDelim(writer, contentLine, file)
	}
	writer.Flush()
}

func writeWithDelim(writer io.Writer, data []string, fileName string) {
	writer.Write([]byte("file: " + fileName))
	writer.Write([]byte{'\n'})
	writer.Write([]byte("_______________________________________________"))
	writer.Write([]byte{'\n'})
	for _, value := range data {
		writer.Write([]byte(value))
		writer.Write([]byte{'\n'})
	}
	writer.Write([]byte("################################################"))
	writer.Write([]byte{'\n'})
}
