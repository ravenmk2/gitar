package data

import (
	"testing"
)

func TestDataStore(t *testing.T) {
	store, err := OpenDataStore(t.TempDir())
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer func() {
		if err := store.Close(); err != nil {
			t.Errorf("close: %v", err)
		}
	}()

	t.Run("github repo", func(t *testing.T) {
		if exists, err := store.GithubRepoExists("o", "r"); err != nil || exists {
			t.Fatalf("exists = %v, err = %v", exists, err)
		}
		if err := store.SaveGithubRepo("o", "r"); err != nil {
			t.Fatalf("save: %v", err)
		}
		if exists, err := store.GithubRepoExists("o", "r"); err != nil || !exists {
			t.Fatalf("exists = %v, err = %v", exists, err)
		}
		if err := store.SaveGithubRepo("o", "r"); err != nil {
			t.Fatalf("save again: %v", err)
		}
	})

	t.Run("gitee repo", func(t *testing.T) {
		if err := store.SaveGiteeRepo("o", "r"); err != nil {
			t.Fatalf("save: %v", err)
		}
		if exists, err := store.GiteeRepoExists("o", "r"); err != nil || !exists {
			t.Fatalf("exists = %v, err = %v", exists, err)
		}
	})

	t.Run("git repo", func(t *testing.T) {
		if err := store.SaveRepo("https://github.com/o/r"); err != nil {
			t.Fatalf("save: %v", err)
		}
		if exists, err := store.RepoExists("https://github.com/o/r"); err != nil || !exists {
			t.Fatalf("exists = %v, err = %v", exists, err)
		}
	})

	t.Run("commit downloaded", func(t *testing.T) {
		if done, err := store.IsCommitDownloaded("abc"); err != nil || done {
			t.Fatalf("done = %v, err = %v", done, err)
		}
		if err := store.SetCommitDownloaded("abc"); err != nil {
			t.Fatalf("set: %v", err)
		}
		if done, err := store.IsCommitDownloaded("abc"); err != nil || !done {
			t.Fatalf("done = %v, err = %v", done, err)
		}
		if err := store.SetCommitDownloaded("abc"); err != nil {
			t.Fatalf("set again: %v", err)
		}
	})

	t.Run("commit mailed", func(t *testing.T) {
		if done, err := store.IsCommitMailed("abc"); err != nil || done {
			t.Fatalf("done = %v, err = %v", done, err)
		}
		if err := store.SetCommitMailed("abc"); err != nil {
			t.Fatalf("set: %v", err)
		}
		if done, err := store.IsCommitMailed("abc"); err != nil || !done {
			t.Fatalf("done = %v, err = %v", done, err)
		}
	})
}
