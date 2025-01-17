package utils

import (
	"os"
)

func EnsureDirectoryExists(directory string, perm os.FileMode) error {
	_, err := os.Stat(directory)
	if err == nil {
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	return os.MkdirAll(directory, perm)
}
