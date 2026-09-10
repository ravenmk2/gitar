package github

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gitar/pkg/client/common"
	"gitar/pkg/utils"
	"github.com/google/go-github/v56/github"
)

type GitHubService struct {
	client *github.Client
}

func NewGitHubService(token string) *GitHubService {
	client := github.NewClient(nil)
	if token != "" {
		client = client.WithAuthToken(token)
	}
	return &GitHubService{
		client: client,
	}
}

func (me *GitHubService) ResolveArchive(url common.RepoUrl) (*common.ArchiveInfo, error) {
	tagName := ""
	if len(url.Release) > 0 {
		tagName = url.Release
	}
	if len(url.Tag) > 0 {
		tagName = url.Tag
	}

	if len(tagName) > 0 {
		return me.resolveArchiveByTag(url, tagName)
	}
	if len(url.Branch) > 0 {
		return me.resolveArchiveByBranch(url)
	}
	if len(url.Commit) > 0 {
		return me.resolveArchiveByCommit(url)
	}

	release, err := me.findBestRelease(url.Owner, url.Repo)
	if err != nil {
		return nil, err
	}
	if release != nil {
		return me.resolveArchiveByTag(url, *release.TagName)
	}

	branch, err := me.findBestBranch(url.Owner, url.Repo)
	if err != nil {
		return nil, err
	}
	if branch != nil {
		return me.branchToArchive(url, branch)
	}

	return nil, errors.New("could not resolve archive")
}

func (me *GitHubService) branchToArchive(url common.RepoUrl, branch *github.Branch) (*common.ArchiveInfo, error) {
	arc := &common.ArchiveInfo{
		Platform: Platform,
	}

	// https://github.com/{owner}/{repo}/archive/{commit-sha}.{format}
	// 使用 Commit ID 保证在下载时和 API 查到的保持一致

	commit := branch.GetCommit().GetSHA()
	arcUrl := fmt.Sprintf("https://github.com/%s/%s/archive/%s", url.Owner, url.Repo, commit)

	arc.Name = fmt.Sprintf("%s-%s-%s", url.Repo, branch.GetName(), shortCommit(commit))
	arc.Name = strings.ReplaceAll(arc.Name, "/", "-")
	arc.Commit = commit
	arc.TarUrl = arcUrl + ".tar.gz"
	arc.ZipUrl = arcUrl + ".zip"

	return validateArchive(arc)
}

func (me *GitHubService) resolveArchiveByCommit(url common.RepoUrl) (*common.ArchiveInfo, error) {
	arc := &common.ArchiveInfo{
		Platform: Platform,
	}

	arcUrl := fmt.Sprintf("https://github.com/%s/%s/archive/%s", url.Owner, url.Repo, url.Commit)

	arc.Name = fmt.Sprintf("%s-%s", url.Repo, shortCommit(url.Commit))
	arc.Commit = url.Commit
	arc.TarUrl = arcUrl + ".tar.gz"
	arc.ZipUrl = arcUrl + ".zip"

	return validateArchive(arc)
}

func (me *GitHubService) resolveArchiveByTag(url common.RepoUrl, tagName string) (*common.ArchiveInfo, error) {
	arc := &common.ArchiveInfo{
		Platform: Platform,
	}

	tag, err := me.findTag(url.Owner, url.Repo, tagName)
	if err != nil {
		return nil, err
	}
	if tag == nil {
		return nil, errors.New("no matched tag")
	}

	// https://github.com/{owner}/{repo}/archive/refs/tags/{tag}.{format}
	// 这里使用 Archive URL 而不使用 REST API 返回的 URL 可以得到更友好的文件名

	arcUrl := fmt.Sprintf("https://github.com/%s/%s/archive/refs/tags/%s", url.Owner, url.Repo, tagName)
	arcName := tagName
	if !strings.HasPrefix(tagName, url.Repo) {
		arcName = fmt.Sprintf("%s-%s", url.Repo, tagName)
	}

	arc.Name = arcName
	arc.Name = strings.ReplaceAll(arc.Name, "/", "-")
	arc.Commit = tag.GetCommit().GetSHA()
	arc.TarUrl = arcUrl + ".tar.gz"
	arc.ZipUrl = arcUrl + ".zip"

	return validateArchive(arc)
}

func (me *GitHubService) resolveArchiveByBranch(url common.RepoUrl) (*common.ArchiveInfo, error) {
	branch, err := me.findBranch(url.Owner, url.Repo, url.Branch)
	if err != nil {
		return nil, err
	}
	if branch == nil {
		return nil, errors.New("no matched branch")
	}

	return me.branchToArchive(url, branch)
}

func (me *GitHubService) findBestBranch(owner, repo string) (*github.Branch, error) {
	desired := utils.NewStringSet([]string{"master", "main", "trunk", "release", "develop"})
	branches := []*github.Branch{}

	for page := 1; page < 100; page++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		opts := &github.BranchListOptions{
			ListOptions: github.ListOptions{Page: page, PerPage: 100},
		}
		items, _, err := me.client.Repositories.ListBranches(ctx, owner, repo, opts)
		cancel()
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		for _, item := range items {
			if desired.Contains(item.GetName()) {
				return item, nil
			}
			branches = append(branches, item)
		}
	}

	if len(branches) == 0 {
		return nil, nil
	}

	for _, branch := range branches {
		name := branch.GetName()
		if strings.HasPrefix(name, "release/") ||
			strings.HasPrefix(name, "release-") {
			return branch, nil
		}
	}

	return branches[0], nil
}

func (me *GitHubService) findAnyWellKnownBranch(owner, repo string) (*github.Branch, error) {
	desired := utils.NewStringSet([]string{"master", "main", "trunk", "release", "develop"})

	for page := 1; page < 100; page++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		opts := &github.BranchListOptions{
			ListOptions: github.ListOptions{Page: page, PerPage: 100},
		}
		branches, _, err := me.client.Repositories.ListBranches(ctx, owner, repo, opts)
		cancel()
		if err != nil {
			return nil, err
		}
		if len(branches) == 0 {
			return nil, errors.New("no branch fetched")
		}
		for _, branch := range branches {
			if desired.Contains(branch.GetName()) {
				return branch, nil
			}
		}
	}

	return nil, errors.New("no matched branch")
}

func (me *GitHubService) findBestRelease(owner, repo string) (*github.RepositoryRelease, error) {
	releases := []*github.RepositoryRelease{}

	for page := 1; page < 100; page++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		opts := &github.ListOptions{
			Page:    page,
			PerPage: 100,
		}
		items, _, err := me.client.Repositories.ListReleases(ctx, owner, repo, opts)
		cancel()
		if err != nil {
			return nil, err
		}
		if len(items) == 0 {
			break
		}

		for _, item := range items {
			if !item.GetDraft() && !item.GetPrerelease() {
				return item, nil
			}
			releases = append(releases, item)
		}
	}

	if len(releases) == 0 {
		return nil, nil
	}

	for _, release := range releases {
		if !release.GetPrerelease() {
			return release, nil
		}
	}

	return releases[0], nil
}

func (me *GitHubService) findAnyRelease(owner, repo string) (*github.RepositoryRelease, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	releases, _, err := me.client.Repositories.ListReleases(ctx, owner, repo, nil)
	if err != nil {
		return nil, err
	}
	if len(releases) == 0 {
		return nil, nil
	}
	return releases[0], nil
}

func (me *GitHubService) findBestTag(owner, repo string) (*github.RepositoryTag, error) {
	for page := 1; page < 100; page++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		opts := &github.ListOptions{Page: page, PerPage: 100}
		tags, _, err := me.client.Repositories.ListTags(ctx, owner, repo, opts)
		cancel()
		if err != nil {
			return nil, err
		}

		if len(tags) == 0 {
			break
		}

		for _, tag := range tags {
			name := tag.GetName()
			if strings.HasPrefix(name, "release/") ||
				strings.HasPrefix(name, "release-") {
				return tag, nil
			}
		}
	}

	return nil, nil
}

func (me *GitHubService) findBranch(owner, repo, name string) (*github.Branch, error) {
	for page := 1; page < 100; page++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		opts := &github.BranchListOptions{
			ListOptions: github.ListOptions{Page: page, PerPage: 100},
		}
		branches, _, err := me.client.Repositories.ListBranches(ctx, owner, repo, opts)
		cancel()
		if err != nil {
			return nil, err
		}

		if len(branches) == 0 {
			return nil, errors.New("no branch fetched")
		}
		for _, branch := range branches {
			if branch.GetName() == name {
				return branch, nil
			}
		}
	}
	return nil, errors.New("no matched branch")
}

func (me *GitHubService) findTag(owner, repo, name string) (*github.RepositoryTag, error) {
	for page := 1; page < 100; page++ {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
		opts := &github.ListOptions{Page: page, PerPage: 100}
		tags, _, err := me.client.Repositories.ListTags(ctx, owner, repo, opts)
		cancel()
		if err != nil {
			return nil, err
		}

		if len(tags) == 0 {
			return nil, errors.New("no tag fetched")
		}

		for _, tag := range tags {
			if tag.GetName() == name {
				return tag, nil
			}
		}
	}

	return nil, errors.New("no matched tag")
}

func shortCommit(commit string) string {
	if len(commit) > 7 {
		return commit[:7]
	}
	return commit
}

func validateArchive(arc *common.ArchiveInfo) (*common.ArchiveInfo, error) {
	if arc.Commit == "" {
		return nil, errors.New("no commit found")
	}
	return arc, nil
}
