package docker

import (
	"archive/tar"
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

	"github.com/kubitre/diplom/models"
	"github.com/kubitre/diplom/utils"
	log "github.com/sirupsen/logrus"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/jsonmessage"
	"github.com/docker/docker/pkg/term"
)

type (
	DockerExecutor struct {
		Status       chan bool // syncronization for executor
		DockerClient *client.Client
	}
)

const (
	buildContextPath  = "dockerBuildContext"
	dockerFileMemName = "Dockerfile"
	entryScript       = "/entry.bash"
)

// Initialize new docker client executor
func NewDockerExecutor() (*DockerExecutor, error) {
	cli, err := client.NewClientWithOpts()
	if err != nil {
		log.Error("can not initiate client for docker api. Error: ", err.Error())
		return nil, err
	}
	return &DockerExecutor{
		Status:       make(chan bool, 1),
		DockerClient: cli,
	}, nil
}

// pulling image from hub.docker.com
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
func (docker *DockerExecutor) CreateContainer(payload *models.ContainerCreatePayload) error {
	ctx := context.Background()
	repsCreating, err := docker.DockerClient.ContainerCreate(ctx, &container.Config{
		Image:      payload.BaseImageName,
		WorkingDir: payload.WorkDir,
		Shell:      payload.ShellCommands,
	}, &container.HostConfig{
		AutoRemove: true,
	}, &network.NetworkingConfig{}, payload.ContainerName)
	if err != nil {
		log.Error("can not create container with default configuration. Error: ", err.Error())
		return docker.removeContainer("test_env_candidate_1")
	}
	log.Debug("success create container. Output oprts: ", repsCreating)
	return nil
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

func (docker *DockerExecutor) PrepareDockerEnv(neededPath map[string]string, dockerFile, shell []string) error {
	fromDockerfile := neededPath
	dockerFile = docker.getPathNeededToCopyInContext(dockerFile, &fromDockerfile)

	dockerF2, err := docker.preparingContext(fromDockerfile, dockerFile, false)
	if err != nil {
		log.Error("can not preparing context from neededpath: ", err)
		return err
	}
	if err := docker.prepareExecutingScript(shell); err != nil {
		log.Error("can not create executing script. ", err)
		return err
	}

	dockerF2 = append(dockerF2, "COPY "+buildContextPath+entryScript+" .")
	dockerF2 = append(dockerF2, "ENTRYPOINT [ \"bash\", \""+entryScript+"\" ]")
	log.Info("result dockerfile: ", dockerF2)

	if err := docker.writeDockerfile(buildContextPath+"/"+dockerFileMemName, docker.preparingBytesFromDockerfile(dockerF2)); err != nil {
		log.Error("can not write dockerfile in buildcontext path. ", err)
		return err
	}
	return nil
}

func (docker *DockerExecutor) preparingContext(neededPath map[string]string, dockerFile []string, fromDockerfile bool) ([]string, error) {
	dockerf := dockerFile
	for key, val := range neededPath {
		lastPart, err := docker.getFinalNamePath(key)
		if err != nil {
			log.Error("can not copy value. ", err)
		}
		if err := docker.copyDir(key, buildContextPath+"/"+lastPart); err != nil {
			return nil, err
		}
		if fromDockerfile {
			dockerf = append(dockerf, "COPY "+buildContextPath+"/"+lastPart+" "+val)
		}
	}
	return dockerf, nil
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
	result, err := utils.CreateExecutingScript(shell)
	if err != nil {
		return err
	}

	if err := ioutil.WriteFile("./"+buildContextPath+entryScript, result, 0777); err != nil {
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

func (docker *DockerExecutor) CreateImageMem(dockerFile, shell, tags []string, neededPath map[string]string) error {
	ctx := context.Background()
	err := docker.PrepareDockerEnv(neededPath, dockerFile, shell)
	if err != nil {
		log.Error("can not readed bytes from fs. " + err.Error())
		return err
	}

	resultBuffer, errCreate := docker.tar(buildContextPath)
	if errCreate != nil {
		log.Error("can not create tar for build context. ", errCreate)
		return err
	}

	dockerFileTar := bytes.NewReader(resultBuffer.Bytes())
	buildOptions := types.ImageBuildOptions{
		Context:    dockerFileTar,
		Dockerfile: buildContextPath + "/" + dockerFileMemName,
		Tags:       tags,
	}
	resp, err := docker.DockerClient.ImageBuild(ctx, dockerFileTar, buildOptions)
	if err != nil {
		log.Error("error while build image by dockerfile. Error: ", err.Error())
		return err
	}
	log.Debug("response from building image: ", resp)
	termFd, isTerm := term.GetFdInfo(os.Stderr)
	if err1 := jsonmessage.DisplayJSONMessagesStream(resp.Body, os.Stderr, termFd, isTerm, nil); err1 != nil {
		return err1
	}
	os.RemoveAll(buildContextPath)

	return nil
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

func (docker *DockerExecutor) removeContainer(containerName string) error {
	ctx := context.Background()
	if err := docker.DockerClient.ContainerRemove(ctx, containerName, types.ContainerRemoveOptions{}); err != nil {
		log.Error("can not remove container. " + err.Error())
		return err
	}
	return nil
}
