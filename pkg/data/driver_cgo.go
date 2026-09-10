//go:build cgo

package data

import (
	_ "github.com/mattn/go-sqlite3"
)

const driverName = "sqlite3"
