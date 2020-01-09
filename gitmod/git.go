package gitmod

import (
	"log"
	"os"

	"gopkg.in/src-d/go-git.v4"
)

type Git struct {
}

func (gt *Git) CloneRepo(url string) error {
	res, err := git.PlainClone("./temp/repo", false, &git.CloneOptions{
		URL:      url,
		Progress: os.Stdout,
	})
	if err != nil {
		return err
	}
	log.Println("result cloning: ", res)
	return nil
}
