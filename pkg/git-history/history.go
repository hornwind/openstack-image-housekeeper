package githistory

import (
	"os"

	"github.com/go-git/go-git/v5"
	log "github.com/hornwind/openstack-image-keeper/pkg/logging"
)

func GetNCommitsFromHead(scanDepth int) ([]string, error) {
	log := log.GetLogger()
	var output []string

	pwd, err := os.Getwd()
	if err != nil {
		log.Debug(err)
		return make([]string, 0), err
	}

	repo, err := git.PlainOpen(pwd)
	if err != nil {
		log.Debug(err)
		return make([]string, 0), err
	}

	logOpts := &git.LogOptions{}

	iter, _ := repo.Log(logOpts)

	for i := 0; i < scanDepth; i++ {
		if commit, err := iter.Next(); commit != nil {
			if err != nil {
				log.Debug(err)
				return make([]string, 0), err
			}
			output = append(output, commit.Hash.String())
		}
	}

	return output, nil
}

func GetCurrentBranch() (string, error) {
	log := log.GetLogger()
	pwd, err := os.Getwd()
	if err != nil {
		log.Debug(err)
		return "", err
	}

	repo, err := git.PlainOpen(pwd)
	if err != nil {
		log.Debug(err)
		return "", err
	}

	ref, err := repo.Head()
	if err != nil {
		log.Debug(err)
		return "", err
	}

	return ref.Name().Short(), nil
}

func GetTags(scanDepth int) ([]string, error) {
	log := log.GetLogger()
	var output []string

	pwd, err := os.Getwd()
	if err != nil {
		log.Debug(err)
		return make([]string, 0), err
	}

	repo, err := git.PlainOpen(pwd)
	if err != nil {
		log.Debug(err)
		return make([]string, 0), err
	}

	iter, _ := repo.Tags()

	for i := 0; i < scanDepth; i++ {
		if ref, err := iter.Next(); ref != nil {
			if err != nil {
				log.Debug(err)
				return make([]string, 0), err
			}
			output = append(output, ref.Name().Short())
		}
	}

	return output, nil
}
