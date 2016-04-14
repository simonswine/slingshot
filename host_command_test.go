package main

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
