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

	if err = os.MkdirAll(cfg.Paths.Temp, os.ModePerm); err != nil {
		return err
	}
	if err = os.MkdirAll(cfg.Paths.Data, os.ModePerm); err != nil {
		return err
	}
	if err = os.MkdirAll(cfg.Paths.Repository, os.ModePerm); err != nil {
		return err
	}

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

// mirrorRefSpec 使本地引用布局与远端完全一致（含 tags 等所有 refs），等价于 git clone --mirror
const mirrorRefSpec = gitcfg.RefSpec("+refs/*:refs/*")

func ensureRemote(repo *git.Repository, url string) error {
	remoteName := "origin"
	mirrorConfig := &gitcfg.RemoteConfig{
		Name:  remoteName,
		URLs:  []string{url},
		Fetch: []gitcfg.RefSpec{mirrorRefSpec},
	}

	remote, err := repo.Remote(remoteName)
	if err != nil {
		if errors.Is(err, git.ErrRemoteNotFound) {
			logrus.Infof("Add remote: %s => %s", remoteName, url)
			_, err = repo.CreateRemote(mirrorConfig)
			return err
		}
		return err
	}

	if remoteConfigMatches(remote.Config(), mirrorConfig) {
		return nil
	}

	logrus.Infof("Update remote: %s => %s", remoteName, url)
	err = repo.DeleteRemote(remoteName)
	if err != nil {
		return err
	}
	_, err = repo.CreateRemote(mirrorConfig)
	return err
}

func remoteConfigMatches(current, expected *gitcfg.RemoteConfig) bool {
	if len(current.URLs) == 0 || current.URLs[0] != expected.URLs[0] {
		return false
	}
	if len(current.Fetch) != len(expected.Fetch) {
		return false
	}
	for i, spec := range expected.Fetch {
		if current.Fetch[i] != spec {
			return false
		}
	}
	return true
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
		Prune:    true,
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
	cmd := exec.Command("git", "fetch", "--prune", remoteName)
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
		err = removeFileWithRetry(file, 5, 200*time.Millisecond)
		if err != nil {
			logrus.Error(err)
		}
	}
}

// removeFileWithRetry 带重试地删除文件。
// Windows 上 fetch 失败后 go-git 打开的 pack 临时文件句柄可能尚未释放（或被杀软扫描占用），
// 直接删除会报 "being used by another process"，稍候重试通常即可成功。
func removeFileWithRetry(path string, attempts int, delay time.Duration) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = os.Remove(path)
		if err == nil || os.IsNotExist(err) {
			return nil
		}
		if i+1 < attempts {
			time.Sleep(delay)
		}
	}
	return err
}
