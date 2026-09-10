package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
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

func MirrorRepository(url string, useSSH, shouldSendMail bool, maxRetries int) error {
	err := DoMirrorRepository(url, useSSH, shouldSendMail, maxRetries)
	if err == nil {
		logrus.Infof("All done")
	}
	return err
}

func DoMirrorRepository(url string, useSSH, shouldSendMail bool, maxRetries int) error {
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
	defer func() {
		if err := store.Close(); err != nil {
			logrus.Error(err)
		}
	}()

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
	repo, err := openOrInit(repoDir)
	if err != nil {
		return err
	}

	err = ensureRemote(repo, gitUrl)
	if err != nil {
		return err
	}

	return repoFetchUntilOk(repoDir, repo, useSSH, maxRetries)
}

func ensureRemote(repo *git.Repository, url string) error {
	remoteName := "origin"
	remote, err := repo.Remote(remoteName)

	if err != nil {
		if errors.Is(err, git.ErrRemoteNotFound) {
			logrus.Infof("Add remote: %s => %s", remoteName, url)
			_, err = repo.CreateRemote(&gitcfg.RemoteConfig{
				Name: remoteName,
				URLs: []string{url},
			})
			return err
		}
		return err
	}

	remoteUrl := remote.Config().URLs[0]
	if remoteUrl == url {
		return nil
	}

	logrus.Infof("Update remote: %s => %s", remoteName, url)
	err = repo.DeleteRemote(remoteName)
	if err != nil {
		return err
	}
	_, err = repo.CreateRemote(&gitcfg.RemoteConfig{
		Name: remoteName,
		URLs: []string{url},
	})
	return err
}

func openOrInit(repoDir string) (*git.Repository, error) {
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
	return git.PlainInitWithOptions(repoDir, &git.PlainInitOptions{
		Bare: true,
	})
}

func repoFetchUntilOk(repoDir string, repo *git.Repository, useSSH bool, maxAttempts int) error {
	remotes, err := repo.Remotes()
	if err != nil {
		return err
	}

	fetchFn := repoFetchBuiltin
	if useSSH {
		fetchFn = repoFetchCli
	}

	if maxAttempts < 1 {
		maxAttempts = 1
	}
	for _, remote := range remotes {
		for attempt := 1; attempt <= maxAttempts; attempt++ {
			err := fetchFn(repoDir, remote)
			repoRemoveTempFiles(repoDir)
			if err == nil {
				return nil
			}
			logrus.Error(err)
			if attempt < maxAttempts {
				delay := time.Duration(attempt*attempt) * time.Second
				logrus.Infof("Retry fetch after %s (%d/%d)", delay, attempt, maxAttempts)
				time.Sleep(delay)
			}
		}
	}

	return errors.New("failed to fetch")
}

func repoFetchBuiltin(repoDir string, remote *git.Remote) error {
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

func repoFetchCli(repoDir string, remote *git.Remote) error {
	remoteName := remote.Config().Name
	logrus.Infof("Fetching %s", remoteName)
	cmd := exec.Command("git", "fetch", remoteName)
	cmd.Dir = repoDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
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
