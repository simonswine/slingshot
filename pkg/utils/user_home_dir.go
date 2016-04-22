package utils

import (
	"path"
)

func SlinshotConfigDirPath() (string, error) {
	homeDir, err := UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homeDir, ".slingshot"), nil
}

func VagrantKeyPath() (string, error) {
	homeDir, err := UserHomeDir()
	if err != nil {
		return "", err
	}
	return path.Join(homeDir, ".vagrant.d/insecure_private_key"), nil
}
