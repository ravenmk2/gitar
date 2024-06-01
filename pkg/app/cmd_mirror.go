package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gitar/pkg/client"
	"gitar/pkg/config"
	"gitar/pkg/data"
	"gitar/pkg/fslock"
	"gitar/pkg/utils"
	"github.com/go-git/go-git/v5"
	gitcfg "github.com/go-git/go-git/v5/config"
	"github.com/sirupsen/logrus"
)

func MirrorRepository(url string, useSSH, shouldSendMail bool) error {
	err := DoMirrorRepository(url, useSSH, shouldSendMail)
	if err == nil {
		logrus.Infof("All done")
	}
	return err
}

func DoMirrorRepository(url string, useSSH, shouldSendMail bool) error {
	logrus.Infof("Mirroring repository: %s", url)

	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	logrus.Infof("Paths: %+v", cfg.Paths)

	repoUrl, err := client.ParseRepoUrl(url)
	if err != nil {
		return err
	}
	logrus.Infof("Platform: %s", repoUrl.Platform)
	logrus.Infof("Repository: %s/%s", repoUrl.Owner, repoUrl.Repo)

	store, err := data.OpenDataStore(cfg.Paths.Data)
	if err != nil {
		return err
	}

	err = SaveRepo(store, repoUrl)
	if err != nil {
		return err
	}

	repoDir := filepath.Join(cfg.Paths.Repository, repoUrl.Platform, repoUrl.Owner, repoUrl.Repo)
	logrus.Infof("Repository directory: %s", repoDir)

	lockName := fmt.Sprintf("%s_%s_%s.lock", repoUrl.Platform, repoUrl.Owner, repoUrl.Repo)
	lockFile := filepath.Join(cfg.Paths.Temp, lockName)
	lock := fslock.New(lockFile)
	err = lock.TryLock()
	if err != nil {
		return err
	}
	defer func(lock fslock.Lock) {
		err := lock.Unlock()
		if err != nil {
			logrus.Error(err)
		}
	}(lock)

	urlFormat := "https://%s/%s/%s.git"
	if useSSH {
		urlFormat = "git@%s:%s/%s.git"
	}
	gitUrl := fmt.Sprintf(urlFormat, repoUrl.Host, repoUrl.Owner, repoUrl.Repo)
	repo, err := openOrInit(repoDir, gitUrl)
	if err != nil {
		return err
	}

	return repoFetchUntilOk(repoDir, repo)
}

func openOrInit(repoDir string, url string) (*git.Repository, error) {
	exists, err := utils.DirExists(repoDir)
	if err != nil {
		return nil, err
	}

	if exists {
		return git.PlainOpenWithOptions(repoDir, &git.PlainOpenOptions{
			DetectDotGit: false,
		})
	}

	logrus.Infof("Initialize local repository")
	repo, err := git.PlainInitWithOptions(repoDir, &git.PlainInitOptions{
		Bare: true,
	})
	if err != nil {
		return nil, err
	}

	_, err = repo.CreateRemote(&gitcfg.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		return nil, err
	}
	return repo, nil
}

func repoFetchUntilOk(repoDir string, repo *git.Repository) error {
	remotes, err := repo.Remotes()
	if err != nil {
		return err
	}

	for _, remote := range remotes {
		for true {
			err := repoFetch(remote)
			repoRemoveTempFiles(repoDir)
			if err == nil {
				return nil
			} else {
				logrus.Error(err)
			}
		}
	}

	return errors.New("failed to fetch")
}

func repoFetch(remote *git.Remote) error {
	logrus.Infof("Fetching %s", remote.Config().Name)
	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	err := remote.FetchContext(ctx, &git.FetchOptions{
		Progress: os.Stdout,
	})
	if err == nil {
		return nil
	}
	if errors.Is(err, git.NoErrAlreadyUpToDate) {
		logrus.Infof("Already up-to-date")
		return nil
	}
	return err
}

func repoRemoveTempFiles(dir string) {
	pattern := dir + "/*/*/tmp_*"
	files, err := filepath.Glob(pattern)
	if err != nil {
		logrus.Error(err)
		return
	}
	if files == nil {
		return
	}
	for _, file := range files {
		rel, err := filepath.Rel(dir, file)
		if err != nil {
			logrus.Error(err)
			continue
		}
		logrus.Warnf("Remove temp file: %s", rel)
		err = os.Remove(file)
		if err != nil {
			logrus.Error(err)
		}
	}
}
