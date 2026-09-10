package fslock

import (
	"errors"
	"fmt"
	"os"

	"golang.org/x/sys/windows"
)

type fsLock struct {
	filename string
	file     *os.File
}

func New(filename string) Lock {
	return &fsLock{filename: filename}
}

func (l *fsLock) Lock() error {
	return l.lock(windows.LOCKFILE_EXCLUSIVE_LOCK)
}

func (l *fsLock) TryLock() error {
	err := l.lock(windows.LOCKFILE_EXCLUSIVE_LOCK | windows.LOCKFILE_FAIL_IMMEDIATELY)
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return fmt.Errorf("locked: %s", l.filename)
	}
	return err
}

func (l *fsLock) lock(flags uint32) error {
	file, err := os.OpenFile(l.filename, os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return err
	}
	err = windows.LockFileEx(windows.Handle(file.Fd()), flags, 0, 1, 0, new(windows.Overlapped))
	if err != nil {
		_ = file.Close()
		return err
	}
	l.file = file
	return nil
}

func (l *fsLock) Unlock() error {
	if l.file == nil {
		return nil
	}
	err := windows.UnlockFileEx(windows.Handle(l.file.Fd()), 0, 1, 0, new(windows.Overlapped))
	cerr := l.file.Close()
	l.file = nil
	if err != nil {
		return err
	}
	if cerr != nil {
		return cerr
	}
	return os.Remove(l.filename)
}
