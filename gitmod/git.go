package gitmod

import (
	log "github.com/sirupsen/logrus"
	"os"
	"strings"

	"github.com/google/uuid"
	"gopkg.in/src-d/go-git.v4"
)

type Git struct {
}

type ErrorType int

const (
	ErrorExistingRepository ErrorType = iota
	ErrorAuthenticate       ErrorType = iota
	ErrorUnrecognized       ErrorType = iota

	stringErrorExistingRepo  = "repository already exists"
	stringErrorAuthenticated = "need auth"
)

func (gt *Git) CloneRepo(url string) (string, error) {
	id := uuid.New()
	res, err := git.PlainClone("repo_"+id.String(), false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return "", err
	}
	log.Info("result cloning: ", res, "into :", "repo_"+id.String())
	return "repo_" + id.String(), nil
}

//nolint:unused
func (gt *Git) getTypeErrorCode(nameError string) ErrorType {
	switch nameError {
	case stringErrorExistingRepo:
		return ErrorExistingRepository
	case stringErrorAuthenticated:
		return ErrorAuthenticate
	default:
		return ErrorUnrecognized
	}
}

func (gt *Git) GetTypeError(err error) ErrorType {
	for key, val := range map[string]ErrorType{
		stringErrorExistingRepo:  ErrorExistingRepository,
		stringErrorAuthenticated: ErrorAuthenticate} {
		if strings.Contains(err.Error(), key) {
			return val
		}
	}
	return ErrorUnrecognized
}

func (gt *Git) RemoveRepo(repoPath string) error {
	return os.RemoveAll(repoPath)
}
