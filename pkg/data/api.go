package data

type DataStore interface {
	Open() error
	Close() error

	RepoExists(repo string) (bool, error)
	SaveRepo(repo string) error

	GithubRepoExists(owner, repo string) (bool, error)
	SaveGithubRepo(owner, repo string) error

	GiteeRepoExists(owner, repo string) (bool, error)
	SaveGiteeRepo(owner, repo string) error

	IsCommitDownloaded(id string) (bool, error)
	SetCommitDownloaded(id string) error

	IsCommitMailed(id string) (bool, error)
	SetCommitMailed(id string) error
}
