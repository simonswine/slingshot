package utils

import (
	"fmt"
	"os"
)

func EnsureDirectory(path string) error {
	if stat, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			err = os.Mkdir(path, 0700)
			if err != nil {
				return fmt.Errorf("failed to create directory '%s': %s", path, err)
			}
		} else {
			return fmt.Errorf("cannot list directory '%s': %s", path, err)
		}
	} else {
		if !stat.IsDir() {
			return fmt.Errorf("path is not a directory '%s'", path)
		}

	}
	return nil
}
