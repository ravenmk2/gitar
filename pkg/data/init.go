package data

import (
	"path/filepath"
)

func OpenDataStore(dir string) (DataStore, error) {
	store := NewSqlite3DataStore(filepath.Join(dir, "gitar.sqlite"))
	err := store.Open()
	if err != nil {
		return nil, err
	}
	return store, nil
}
