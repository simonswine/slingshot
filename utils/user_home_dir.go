package utils

import (
	"os"
	"runtime"
	"path"
)

func UserHomeDir() string {
	if runtime.GOOS == "windows" {
		home := os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH")
		if home == "" {
			home = os.Getenv("USERPROFILE")
		}
		return home
	}
	return os.Getenv("HOME")
}

func VagrantKeyPath() string {
	return path.Join(UserHomeDir(), ".vagrant.d/insecure_private_key")
}