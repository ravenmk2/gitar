package fslock

import (
	"github.com/sirupsen/logrus"
)

type fsLock struct {
	filename string
}

func New(filename string) Lock {
	logrus.Warnf("Fslock for windows is not supported")
	return &fsLock{filename: filename}
}

func (l *fsLock) Lock() error {
	return nil
}

func (l *fsLock) TryLock() error {
	return nil
}

func (l *fsLock) Unlock() error {
	return nil
}
