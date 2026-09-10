//go:build !cgo

package data

import (
	_ "modernc.org/sqlite"
)

const driverName = "sqlite"
