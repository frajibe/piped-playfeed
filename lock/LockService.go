package lock

import (
	"errors"
	"fmt"
	"os"
	"piped-playfeed/utils"
)

const lockFile = "piped-playfeed.lock"

func CreateLockFile() error {
	if _, err := os.Stat(lockFile); err == nil {
		msg := fmt.Sprintf("'%s' file is present.\n"+
			"- Reason 1: the application is already running -> wait for its end and retry.\n"+
			"- Reason 2: the previous run failed -> check the log file to understand why, and then delete the lock file.", lockFile)
		return errors.New(msg)
	}
	if _, err := os.Create(lockFile); err != nil {
		return utils.WrapError(fmt.Sprintf("invalid location '%s'", lockFile), err)
	}
	return nil
}

func DeleteLockFile() error {
	if err := os.Remove(lockFile); err != nil {
		return err
	}
	return nil
}
