package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"os"
)

func prepareDockerCommand(t *testing.T) *Command {
	// TODO: fix tests fail if the image is not available locally
	if len(os.Getenv("DOCKER_ENABLE")) == 0 {
		t.Skip("Skipping docker integration tests: Set environment DOCKER_ENABLE=true to enable")
	}

	p := &MockProvider{}

	_, err := p.Docker().Info()
	if err != nil {
		t.Skip("No docker available")
	}

	c := &Command{}
	c.Init(
		&CommandConfig{
			Type: "docker",
		},
		&MockProvider{},
	)
	return c
}

func TestDockerCommandExecuteSucceedStdout(t *testing.T) {
	c := prepareDockerCommand(t)

	stdout, stderr, exitCode, err := c.Execute([]string{"echo", "test"})

	assert.Nil(t, err, "Unexpected error during parsing")
	assert.Equal(t, "test\n", stdout)
	assert.Equal(t, "", stderr)
	assert.Equal(t, 0, exitCode)
}

func TestDockerCommandExecuteFailStderr(t *testing.T) {
	c := prepareDockerCommand(t)

	stdout, stderr, exitCode, err := c.Execute([]string{"ls", "/notexisting"})

	assert.Nil(t, err, "Unexpected error during parsing")
	assert.Equal(t, "", stdout)
	assert.Equal(t, "ls: /notexisting: No such file or directory\n", stderr)
	assert.Equal(t, 1, exitCode)
}

func TestDockerCommandPersistence(t *testing.T) {
	c := prepareDockerCommand(t)
	c.commandImplementation.Config().PersistPaths = []string{
		"test.txt",
		"test/",
	}

	_, _, exitCode, err := c.Execute([]string{"/bin/sh", "-c", "echo test987 > test.txt"})
	assert.Nil(t, err, "Unexpected error during execution")
	assert.Equal(t, 0, exitCode)

	_, _, exitCode, err = c.Execute([]string{"/bin/sh", "-c", "mkdir test; echo test654 > test/test.txt"})
	assert.Nil(t, err, "Unexpected error during execution")
	assert.Equal(t, 0, exitCode)

	stdout, _, exitCode, err := c.Execute([]string{"cat", "test.txt"})
	assert.Nil(t, err, "Unexpected error during execution")
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "test987\n", stdout)

	stdout, _, exitCode, err = c.Execute([]string{"cat", "test/test.txt"})
	assert.Nil(t, err, "Unexpected error during execution")
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "test654\n", stdout)

}
