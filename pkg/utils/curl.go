package utils

import (
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

func CurlDownload(url string, dir string, file string, maxTries int) error {
	if maxTries < 0 {
		maxTries = 5
	} else if maxTries < 1 {
		maxTries = 1
	}

	args := []string{
		"--fail",
		"--location",
		"--output",
		file,
		url,
	}

	var err error
	for i := 0; i < maxTries; i++ {
		err = execCurl(dir, args, true)
		if err == nil {
			return nil
		}
		if i+1 < maxTries {
			delay := time.Duration(i+1) * 3 * time.Second
			logrus.Warnf("Download failed, retry after %s (%d/%d)", delay, i+1, maxTries)
			time.Sleep(delay)
		}
	}

	return err
}

func execCurl(dir string, args []string, redirect bool) error {
	cmd := exec.Command("curl", args...)
	cmd.Dir = dir
	if redirect {
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	}
	err := cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Wait()
}
