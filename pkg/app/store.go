package app

import (
	"fmt"

	"gitar/pkg/client/common"
	"gitar/pkg/client/gitee"
	"gitar/pkg/client/github"
	"gitar/pkg/data"
)

func SaveRepo(store data.DataStore, url *common.RepoUrl) error {
	if url.Platform == github.Platform {
		return store.SaveGithubRepo(url.Owner, url.Repo)
	}
	if url.Platform == gitee.Platform {
		return store.SaveGiteeRepo(url.Owner, url.Repo)
	}
	return fmt.Errorf("unsupported platform: %s", url.Platform)
}
