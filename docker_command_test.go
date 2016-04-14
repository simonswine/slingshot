package main

import (
	"testing"

	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

func prepareCommand(t *testing.T) *Command {
	dockerClient, err := docker.NewClientFromEnv()
	if err != nil {
		t.Skip("No docker available")
	}

	_, err = dockerClient.Info()
	if err != nil {
		t.Skip("No docker available")
	}

	c := &Command{
		commandImplementation: &DockerCommand{
			imageId:      "busybox",
			dockerClient: dockerClient,
		},
	}
	return c
}

func TestDockerCommandExecuteSucceedStdout(t *testing.T) {
	c := prepareCommand(t)

	stdout, stderr, exitCode, err := c.Execute([]string{"echo", "test"})

	assert.Nil(t, err, "Unexpected error during parsing")
	assert.Equal(t, "test\n", stdout)
	assert.Equal(t, "", stderr)
	assert.Equal(t, 0, exitCode)
}

func TestDockerCommandExecuteFailStderr(t *testing.T) {
	c := prepareCommand(t)

	stdout, stderr, exitCode, err := c.Execute([]string{"ls", "/notexisting"})

	assert.Nil(t, err, "Unexpected error during parsing")
	assert.Equal(t, "", stdout)
	assert.Equal(t, "ls: /notexisting: No such file or directory\n", stderr)
	assert.Equal(t, 1, exitCode)
}
