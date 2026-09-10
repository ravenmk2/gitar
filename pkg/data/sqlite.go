package data

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/mattn/go-sqlite3"
)

func NewSqlite3DataStore(dsn string) DataStore {
	store := &Sqlite3DataStore{
		dsn: dsn,
		db:  nil,
	}
	return store
}

type Sqlite3DataStore struct {
	dsn string
	db  *sqlx.DB
}

func (me *Sqlite3DataStore) String() string {
	return fmt.Sprintf("Sqlite3DataStore { DataSource: %s }", me.dsn)
}

func (me *Sqlite3DataStore) Open() error {
	db, err := sqlx.Open("sqlite3", me.dsn)
	if err != nil {
		return err
	}

	err = db.Ping()
	if err != nil {
		return err
	}

	me.db = db
	return me.initDatabase()
}

func (me *Sqlite3DataStore) Close() error {
	if me.db == nil {
		return errors.New("db is not open")
	}
	return me.db.Close()
}

func (me *Sqlite3DataStore) initDatabase() error {
	cmd := `
	CREATE TABLE IF NOT EXISTS [git_repo] (
		[repo] TEXT NOT NULL PRIMARY KEY
	);

	CREATE TABLE IF NOT EXISTS [github_repo] (
		[owner] TEXT NOT NULL,
		[repo]  TEXT NOT NULL,
		PRIMARY KEY([owner], [repo])
	);

	CREATE TABLE IF NOT EXISTS [gitee_repo] (
		[owner] TEXT NOT NULL,
		[repo]  TEXT NOT NULL,
		PRIMARY KEY([owner], [repo])
	);

	CREATE TABLE IF NOT EXISTS [commit_downloaded] (
		[id] TEXT NOT NULL PRIMARY KEY
	);

	CREATE TABLE IF NOT EXISTS [commit_mailed] (
		[id] TEXT NOT NULL PRIMARY KEY
	);
	`
	_, err := me.db.Exec(cmd)
	if err != nil {
		return err
	}
	return nil
}

func (me *Sqlite3DataStore) queryExists(q string, args ...any) (bool, error) {
	rows, err := me.db.Query(q, args...)
	if err != nil {
		return false, err
	}

	defer func(rows *sql.Rows) {
		_ = rows.Close()
	}(rows)

	rows.Next()
	var count int32
	err = rows.Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (me *Sqlite3DataStore) queryExistsByKey(table, field string, value any) (bool, error) {
	cmd := fmt.Sprintf("SELECT count(*) FROM [%s] WHERE [%s] = ?;", table, field)
	return me.queryExists(cmd, value)
}

func (me *Sqlite3DataStore) RepoExists(repo string) (bool, error) {
	return me.queryExistsByKey("git_repo", "repo", repo)
}

func (me *Sqlite3DataStore) SaveRepo(repo string) error {
	cmd := "INSERT OR IGNORE INTO [git_repo] ([repo]) VALUES(?);"
	_, err := me.db.Exec(cmd, repo)
	return err
}

func (me *Sqlite3DataStore) GithubRepoExists(owner, repo string) (bool, error) {
	cmd := "SELECT count(*) FROM [github_repo] WHERE [owner] = ? AND [repo] = ?;"
	return me.queryExists(cmd, owner, repo)
}

func (me *Sqlite3DataStore) SaveGithubRepo(owner, repo string) error {
	cmd := "INSERT OR IGNORE INTO [github_repo] ([owner], [repo]) VALUES(?, ?);"
	_, err := me.db.Exec(cmd, owner, repo)
	return err
}

func (me *Sqlite3DataStore) GiteeRepoExists(owner, repo string) (bool, error) {
	cmd := "SELECT count(*) FROM [gitee_repo] WHERE [owner] = ? AND [repo] = ?;"
	return me.queryExists(cmd, owner, repo)
}

func (me *Sqlite3DataStore) SaveGiteeRepo(owner, repo string) error {
	cmd := "INSERT OR IGNORE INTO [gitee_repo] ([owner], [repo]) VALUES(?, ?);"
	_, err := me.db.Exec(cmd, owner, repo)
	return err
}

func (me *Sqlite3DataStore) IsCommitDownloaded(id string) (bool, error) {
	return me.queryExistsByKey("commit_downloaded", "id", id)
}

func (me *Sqlite3DataStore) SetCommitDownloaded(id string) error {
	cmd := "INSERT OR IGNORE INTO [commit_downloaded] ([id]) VALUES(?);"
	_, err := me.db.Exec(cmd, id)
	return err
}

func (me *Sqlite3DataStore) IsCommitMailed(id string) (bool, error) {
	return me.queryExistsByKey("commit_mailed", "id", id)
}

func (me *Sqlite3DataStore) SetCommitMailed(id string) error {
	cmd := "INSERT OR IGNORE INTO [commit_mailed] ([id]) VALUES(?);"
	_, err := me.db.Exec(cmd, id)
	return err
}
