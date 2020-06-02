package docker_runner

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/tools"
	log "github.com/sirupsen/logrus"
)

type (
	/*DockerExecutor - главный исполняющий модуль заданий связанных с докером*/
	DockerExecutor struct {
		Status       chan bool // syncronization for executor
		DockerClient *client.Client
	}
)

const (
	buildContextPath  = "dockerBuildContext"
	dockerFileMemName = "Dockerfile"
	entryScript       = "entry.bash"
)

// NewDockerExecutor - создание нового докер исполнителя
func NewDockerExecutor() (*DockerExecutor, error) {
	cli, err := client.NewEnvClient()
	if err != nil {
		log.Error("can not initiate client for docker api. Error: ", err.Error())
		return nil, err
	}
	return &DockerExecutor{
		Status:       make(chan bool, 1),
		DockerClient: cli,
	}, nil
}

// PullImage - пуллинг публичных образов
func (docker *DockerExecutor) PullImage(image string) error {
	ctx := context.Background()
	respPulling, errPulling := docker.DockerClient.ImagePull(ctx, image, types.ImagePullOptions{})
	if errPulling != nil {
		log.Error("Can npt pulling image. Error: ", errPulling.Error())
		return errPulling
	}
	// io.Copy(os.Stdout, respPulling)
	log.Debug("response from pulling image: ", respPulling)
	return nil
}

// CreateContainer - function for creating new container with docker file such as json
func (docker *DockerExecutor) CreateContainer(payload *models.ContainerCreatePayload) (string, error) {
	ctx := context.Background()

	hostBinding := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: "8000",
	}
	containerPort, err := nat.NewPort("tcp", "80")
	if err != nil {
		panic("Unable to get the port")
	}
	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}
	repsCreating, err := docker.DockerClient.ContainerCreate(ctx, &container.Config{
		Image: payload.BaseImageName,
	}, &container.HostConfig{
		AutoRemove:   false,
		PortBindings: portBinding,
	}, nil, payload.ContainerName)
	if err != nil {
		log.Error("can not create container with default configuration. Error: ", err.Error())
		return "", docker.RemoveContainer(payload.ContainerName)
	}
	log.Info("success create container. Output oprts: ", repsCreating)
	return repsCreating.ID, nil
}

/*RunContainer - запуск контейнера*/
func (docker *DockerExecutor) RunContainer(containerID string, timeout int64) (io.ReadCloser, error) {
	log.Debug("I'm here. Enter timeout: ", timeout)
	resTimeout := int64(50000)
	if timeout > 0 {
		resTimeout = timeout
	}
	log.Debug("Run container for amount ms: ", resTimeout, " ContainerID: ", containerID)
	ctx := context.Background()
	if errStart := docker.DockerClient.ContainerStart(ctx, containerID, types.ContainerStartOptions{}); errStart != nil {
		return nil, errStart
	}
	statusCH, errCh := docker.DockerClient.ContainerWait(ctx, containerID, container.WaitConditionNextExit)
	select {
	case err := <-errCh:
		if err != nil {
			return nil, err
		}
	case <-statusCH:
	case <-time.After(time.Millisecond * time.Duration(resTimeout)):
		log.Error("container can not return reposne for timeout")
		log.Debug("stop container")
		return nil, errors.New("timeout for starting container")
	}
	log.Info("container start: ", containerID)
	result := make(chan io.ReadCloser, 1)
	err := make(chan error, 1)
	go func(containerID string, resultClose chan io.ReadCloser, resultError chan error) {
		response, err := docker.DockerClient.ContainerLogs(ctx, containerID, types.ContainerLogsOptions{
			ShowStdout: true,
			ShowStderr: true,
			Follow:     true,
		})
		if err != nil {
			log.Error("can not reading docker container logs: ", err)
			// return nil, err
			resultError <- err
			return
		}
		resultClose <- response

	}(containerID, result, err)

	select {
	case res := <-result:
		log.Debug("completed reading logs from container")
		return res, nil
	case err := <-err:
		log.Error("something error while handling logs in container: ", err)
		return nil, err
	case <-time.After(time.Millisecond * time.Duration(resTimeout)):
		log.Error("container can not return reposne for timeout")
		log.Debug("stop container")
		if errStop := docker.DockerClient.ContainerStop(ctx, containerID, nil); errStop != nil {
			log.Error("can not stoped container: ", errStop)
		}
		if errRemoveContainer := docker.RemoveContainer(containerID); errRemoveContainer != nil {
			log.Error("can not remove container: ", errRemoveContainer)
		}
		return nil, errors.New("container not answered for timeout")
	}
}

func (docker *DockerExecutor) preparingBytesFromDockerfile(dockerFile []string) []byte {
	result := ""
	for _, v := range dockerFile {
		result += v + "\n"
	}

	return []byte(result)
}

func (docker *DockerExecutor) getPathNeededToCopyInContext(dockerFile []string, neededPath *map[string]string) []string {
	result := make([]string, 0)
	for _, value := range dockerFile {
		if strings.Contains(value, "COPY") {
			tokens := strings.Split(value, " ")
			lastPart, err := docker.getFinalNamePath(tokens[1])
			if err != nil {
				log.Error("can not copy value. ", err)
			}
			result = append(result, "COPY "+buildContextPath+"/"+lastPart+" "+tokens[2])
			(*neededPath)[tokens[1]] = tokens[2]

		} else {
			result = append(result, value)
		}
	}
	return result
}

func (docker *DockerExecutor) getFinalNamePath(path string) (string, error) {
	fileParts := strings.Split(path, "/")
	if len(fileParts) == 0 {
		return "", errors.New("can not get last part of file")
	}
	return fileParts[len(fileParts)-1], nil
}

/*PrepareDockerEnv - подготовка докер файла для его сборки
 */
func (docker *DockerExecutor) PrepareDockerEnv(neededPath map[string]string, dockerFile, shell []string) error {
	fromDockerfile := neededPath
	dockerFile = docker.getPathNeededToCopyInContext(dockerFile, &fromDockerfile)
	log.Println("DockerFile: ", dockerFile)
	dockerF2, err := docker.preparingContext(fromDockerfile, dockerFile, true)
	if err != nil {
		log.Error("can not preparing context from neededpath: ", err)
		return err
	}
	os.Mkdir(buildContextPath, 0777)
	if len(shell) > 0 {
		if err := docker.prepareExecutingScript(shell); err != nil {
			log.Error("can not create executing script. ", err)
			return err
		}

		dockerF2 = append(dockerF2, "COPY "+buildContextPath+"/"+entryScript+" .")
		dockerF2 = append(dockerF2, "ENTRYPOINT [ \"bash\", \""+entryScript+"\" ]")
		log.Info("result dockerfile: ", dockerF2)
	}

	if err := docker.writeDockerfile(buildContextPath+"/"+dockerFileMemName, docker.preparingBytesFromDockerfile(dockerF2)); err != nil {
		log.Error("can not write dockerfile in buildcontext path. ", err)
		return err
	}
	return nil
}

func (docker *DockerExecutor) preparingContext(neededPath map[string]string, dockerFile []string, fromDockerfile bool) ([]string, error) {
	dockerf := dockerFile
	for key, val := range neededPath {
		log.Println("key: ", key, " value: ", val)
		lastPart, err := docker.getFinalNamePath(key)
		log.Println(lastPart)
		if err != nil {
			log.Error("can not copy value. ", err)
		}
		if err := docker.copyDir(key, buildContextPath+"/"+lastPart); err != nil {
			return nil, err
		}
		if fromDockerfile {
			dockerf = docker.findAnnotationRepoCandidate(dockerf, "COPY "+buildContextPath+"/"+lastPart+" "+val)
		}
	}
	log.Println("DOCKERFILE: ", dockerf)
	return dockerf, nil
}

func (docker *DockerExecutor) findAnnotationRepoCandidate(dockerFile []string, repoCandidate string) []string {
	for index, value := range dockerFile {
		switch value {
		case "{{repoCandidate}}":
			dockerFile[index] = repoCandidate
		case "{{workdir repoCandidate}}":
			dockerFile[index] = "WORKDIR repoCandidate"
		}
	}
	return dockerFile
}

func (docker *DockerExecutor) writeDockerfile(filePath string, data []byte) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()
	return ioutil.WriteFile(filePath, data, 0777)
}

func (docker *DockerExecutor) copyDir(src string, dst string) (err error) {
	src = filepath.Clean(src)
	dst = filepath.Clean(dst)
	log.Println("copy from source: ", src, " to destination: ", dst)

	si, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !si.IsDir() {
		return fmt.Errorf("source is not a directory")
	}

	_, err = os.Stat(dst)
	if err != nil && !os.IsNotExist(err) {
		return
	}

	err = os.MkdirAll(dst, si.Mode())
	if err != nil {
		return
	}

	entries, err := ioutil.ReadDir(src)
	if err != nil {
		return
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = docker.copyDir(srcPath, dstPath)
			if err != nil {
				return
			}
		} else {
			// Skip symlinks.
			if entry.Mode()&os.ModeSymlink != 0 {
				continue
			}

			err = docker.copyFile(srcPath, dstPath)
			if err != nil {
				return
			}
		}
	}

	return
}

func (docker *DockerExecutor) prepareExecutingScript(shell []string) error {
	result, err := tools.CreateExecutingScript(shell)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile("./"+buildContextPath+"/"+entryScript, result, 0777); err != nil {
		return err
	}
	return nil
}

func (docker *DockerExecutor) copyFile(src, dst string) (err error) {
	in, err := os.Open(src)
	if err != nil {
		return
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return
	}
	defer func() {
		if e := out.Close(); e != nil {
			err = e
		}
	}()

	_, err = io.Copy(out, in)
	if err != nil {
		return
	}

	err = out.Sync()
	if err != nil {
		return
	}

	si, err := os.Stat(src)
	if err != nil {
		return
	}
	err = os.Chmod(dst, si.Mode())
	if err != nil {
		return
	}

	return
}

func (docker *DockerExecutor) tar(src string) (*bytes.Buffer, error) {
	buff, err := docker.compressDir(buildContextPath)
	if err != nil {
		return nil, err
	}

	return buff, nil
}

func (docker *DockerExecutor) compressDir(path string) (*bytes.Buffer, error) {
	buf := new(bytes.Buffer)
	zr := gzip.NewWriter(buf)
	tw := tar.NewWriter(zr)

	if err1 := filepath.Walk(path, func(file string, fi os.FileInfo, err error) error {
		header, err2 := tar.FileInfoHeader(fi, file)
		if err2 != nil {
			return err2
		}

		header.Name = filepath.ToSlash(file)

		if err2 := tw.WriteHeader(header); err2 != nil {
			return err2
		}
		if !fi.IsDir() {
			data, err2 := os.Open(file)
			if err2 != nil {
				return err2
			}
			if _, err2 := io.Copy(tw, data); err2 != nil {
				return err2
			}
		}
		return nil
	}); err1 != nil {
		return nil, err1
	}

	if err := tw.Close(); err != nil {
		return nil, err
	}
	if err := zr.Close(); err != nil {
		return nil, err
	}
	//
	log.Info("buffer: ", buf)
	return buf, nil
}

/*CreateImageMem - создание образа по заданному dockerfile с заданными инструкциями для выполнения + пометка образа списком тэгов
 */
func (docker *DockerExecutor) CreateImageMem(dockerFile, shell, tags []string, neededPath map[string]string) ([]string, error) {
	ctx := context.Background()
	err := docker.PrepareDockerEnv(neededPath, dockerFile, shell)
	if err != nil {
		log.Error("can not readed bytes from fs. " + err.Error())
		os.RemoveAll(buildContextPath)
		return nil, err
	}

	resultBuffer, errCreate := docker.tar(buildContextPath)
	if errCreate != nil {
		log.Error("can not create tar for build context. ", errCreate)
		os.RemoveAll(buildContextPath)
		return nil, err
	}

	dockerFileTar := bytes.NewReader(resultBuffer.Bytes())
	buildOptions := types.ImageBuildOptions{

		Context:    dockerFileTar,
		Dockerfile: buildContextPath + "/" + dockerFileMemName,
		Tags:       tags,
	}
	log.Debug("TAGS FOR CREATING IMAGE: ", tags)
	resp, err := docker.DockerClient.ImageBuild(ctx, dockerFileTar, buildOptions)
	if err != nil {
		log.Error("error while build image by dockerfile. Error: ", err.Error())
		os.RemoveAll(buildContextPath)
		return nil, err
	}
	log.Debug("response from building image: ", resp)
	// termFd, isTerm := term.GetFdInfo(os.Stderr)
	// if err1 := jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil); err1 != nil {
	// 	return err1
	// }
	os.RemoveAll(buildContextPath)

	return docker.readLogsFromBodyCloser(resp.Body), nil
}

func (docker *DockerExecutor) readLogsFromBodyCloser(rd io.ReadCloser) []string {
	reader := bufio.NewReader(rd)

	result := []string{}
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			} else {
				log.Error(err)
				return nil
			}
		}
		result = append(result, line)
	}
	return result
}

// func (docker *DockerExecutor) CreateImageDockerFile(dockerFilePath string, tags []string) error {
// 	ctx := context.Background()
// 	resultBuffer, err := docker.initiateTarFromFS(dockerFilePath)
// 	if err != nil {
// 		log.Error("can not readed bytes from fs. " + err.Error())
// 		return err
// 	}

// 	dockerFileTar := bytes.NewReader(resultBuffer.Bytes())

// 	buildOptions := types.ImageBuildOptions{
// 		Context:    dockerFileTar,
// 		Dockerfile: dockerFilePath,
// 		Tags:       tags,
// 	}
// 	resp, err := docker.DockerClient.ImageBuild(ctx, dockerFileTar, buildOptions)
// 	if err != nil {
// 		log.Error("error while build image by dockerfile. Error: ", err.Error())
// 		return err
// 	}
// 	log.Debug("response from building image: ", resp)
// 	termFd, isTerm := term.GetFdInfo(os.Stderr)
// 	if err1 := jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil); err1 != nil {
// 		return err1
// 	}
// 	return nil
// }

/*RemoveContainer - удаление контейнера*/
func (docker *DockerExecutor) RemoveContainer(containerName string) error {
	ctx := context.Background()
	if err := docker.DockerClient.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{}); err != nil {
		log.Error("can not remove container. " + err.Error())
		return err
	}
	return nil
}

/*RemoveImage - удаление образа*/
func (docker *DockerExecutor) RemoveImage(imageName string) error {
	ctx := context.Background()
	delResponse, errDelete := docker.DockerClient.ImageRemove(ctx, imageName, types.ImageRemoveOptions{})
	if errDelete != nil {
		return errDelete
	}
	log.Debug(delResponse)
	return nil
}
