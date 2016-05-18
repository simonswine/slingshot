package slingshot

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func prepareHostCommand(t *testing.T) *Command {
	c := &Command{
		commandImplementation: &HostCommand{},
	}
	return c
}

func TestHostCommandExecuteSucceed(t *testing.T) {
	c := prepareHostCommand(t)

	stdout, stderr, exitCode, err := c.Execute([]string{"echo", "test"})

	assert.Nil(t, err, "Unexpected error during parsing")
	assert.Equal(t, "test\n", stdout)
	assert.Equal(t, 0, len(stderr))
	assert.Equal(t, 0, exitCode)
}

func TestHostCommandExecuteFail(t *testing.T) {
	c := prepareHostCommand(t)

	stdout, stderr, exitCode, err := c.Execute([]string{"ls", "/notexisting"})

	assert.Nil(t, err, "Unexpected error during parsing")
	assert.Equal(t, 0, len(stdout))
	assert.True(t, len(stderr) > 0, "Outputs stderr")
	assert.Equal(t, 2, exitCode)
}

func TestHostCommandPersistence(t *testing.T) {
	c := &Command{
		commandImplementation: &HostCommand{
			BaseCommand: BaseCommand{
				config: &CommandConfig{
					PersistPaths: []string{
						"test.txt",
						"test/",
					},
				},
			},
		},
		provider: &MockProvider{},
	}

	_, _, exitCode, err := c.Execute([]string{"/bin/sh", "-c", "echo test987 > test.txt"})
	assert.Nil(t, err, "Unexpected error during execution")
	assert.Equal(t, 0, exitCode)

	_, _, exitCode, err = c.Execute([]string{"/bin/sh", "-c", "mkdir test; echo test654 > test/test.txt"})
	assert.Nil(t, err, "Unexpected error during execution")
	assert.Equal(t, 0, exitCode)

	_, _, exitCode, err = c.Execute([]string{"/bin/sh", "-c", "mkdir test_file; echo test321 > test_file/test_file.txt"})
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

	stdout, _, exitCode, err = c.Execute([]string{"cat", "test_file/test_file.txt"})
	assert.Nil(t, err, "Unexpected error during execution")
	assert.Equal(t, 0, exitCode)
	assert.Equal(t, "test321\n", stdout)

}
