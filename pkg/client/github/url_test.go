package github

import (
	"testing"
)

func TestParseGithubRepoUrl(t *testing.T) {
	commit := "0123456789abcdef0123456789abcdef01234567"
	tests := []struct {
		name   string
		rawUrl string
		owner  string
		repo   string
	}{
		{name: "https", rawUrl: "https://github.com/o/r", owner: "o", repo: "r"},
		{name: "https dot git", rawUrl: "https://github.com/o/r.git", owner: "o", repo: "r"},
		{name: "ssh", rawUrl: "git@github.com:o/r.git", owner: "o", repo: "r"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseGithubRepoUrl(tt.rawUrl)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if info.Platform != Platform || info.Host != Host {
				t.Errorf("platform/host = %s/%s", info.Platform, info.Host)
			}
			if info.Owner != tt.owner || info.Repo != tt.repo {
				t.Errorf("owner/repo = %s/%s, want %s/%s", info.Owner, info.Repo, tt.owner, tt.repo)
			}
		})
	}

	t.Run("release tag", func(t *testing.T) {
		info, err := ParseGithubRepoUrl("https://github.com/o/r/releases/tag/v1.2.3")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Tag != "v1.2.3" || info.Release != "v1.2.3" {
			t.Errorf("tag/release = %s/%s", info.Tag, info.Release)
		}
	})

	t.Run("tree commit", func(t *testing.T) {
		info, err := ParseGithubRepoUrl("https://github.com/o/r/tree/" + commit)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Commit != commit {
			t.Errorf("commit = %s", info.Commit)
		}
	})

	t.Run("tree branch", func(t *testing.T) {
		info, err := ParseGithubRepoUrl("https://github.com/o/r/tree/main")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if info.Branch != "main" || info.RefName != "main" {
			t.Errorf("branch/ref = %s/%s", info.Branch, info.RefName)
		}
	})

	t.Run("unsupported", func(t *testing.T) {
		if _, err := ParseGithubRepoUrl("https://example.com/o/r"); err == nil {
			t.Error("expected error for unsupported url")
		}
	})
}

func TestShortCommit(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"0123456789abcdef0123456789abcdef01234567", "0123456"},
		{"0123456", "0123456"},
		{"abc", "abc"},
		{"", ""},
	}
	for _, tt := range tests {
		if got := shortCommit(tt.in); got != tt.want {
			t.Errorf("shortCommit(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
