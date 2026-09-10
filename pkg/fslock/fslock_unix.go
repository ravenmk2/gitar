//go:build linux || darwin

package fslock

import (
	"fmt"
	"os"
	"syscall"
)

type fsLock struct {
	filename string
	fd       int
}

func New(filename string) Lock {
	return &fsLock{filename: filename}
}

func (l *fsLock) Lock() error {
	if err := l.open(); err != nil {
		return err
	}
	return syscall.Flock(l.fd, syscall.LOCK_EX)
}

func (l *fsLock) TryLock() error {
	if err := l.open(); err != nil {
		return err
	}
	err := syscall.Flock(l.fd, syscall.LOCK_EX|syscall.LOCK_NB)
	if err != nil {
		_ = syscall.Close(l.fd)
	}
	if err == syscall.EWOULDBLOCK {
		return fmt.Errorf("locked: %s", l.filename)
	}
	return err
}

func (l *fsLock) open() error {
	fd, err := syscall.Open(l.filename, syscall.O_CREAT|syscall.O_RDONLY, 0600)
	if err != nil {
		return err
	}
	l.fd = fd
	return nil
}

func (l *fsLock) Unlock() error {
	err := syscall.Close(l.fd)
	if err != nil {
		return err
	}
	return os.Remove(l.filename)
}
