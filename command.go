package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

type CommandInterface interface {
	Exec(command []string, stdout io.Writer, stderr io.Writer, stdin io.Reader) (exitCode int, err error)
	Prepare(*[]byte) error
	Output() ([]byte, error)
	CleanUp()
	Config() *CommandConfig
}

type CommandConfig struct {
	ParameterFile     *string    `yaml:"parameterFile"`
	ResultFile        *string    `yaml:"resultFile"`
	PersistPaths      []string   `yaml:"persistPaths"`
	Type              string     `yaml:"type"`
	WorkingDirContent string     `yaml:"workingDirContent"`
	Execs             [][]string `yaml:"execs"`
}

type Command struct {
	commandImplementation CommandInterface
}

func NewCommand(c *CommandConfig, p *Provider) (*Command, error) {
	cmd := &Command{}
	err := cmd.Init(c, p)
	return cmd, err
}

func (c *Command) Init(config *CommandConfig, p *Provider) error {

	if config.Type == "host" {
		c.commandImplementation = &HostCommand{
			BaseCommand: BaseCommand{
				config: config,
			},
		}
	} else if config.Type == "docker" {
		c.commandImplementation = &DockerCommand{
			BaseCommand: BaseCommand{
				config: config,
			},
			dockerClient: p.docker,
			imageId:      *p.imageId,
		}
	} else {
		return fmt.Errorf("command type '%s' not found", config.Type)
	}

	return nil
}

func (c *Command) Execute(command []string) (stdOut string, stdErr string, exitCode int, err error) {
	err = c.commandImplementation.Prepare(nil)
	if err != nil {
		return
	}
	defer c.commandImplementation.CleanUp()

	var bufOut bytes.Buffer
	var bufErr bytes.Buffer

	exitCode, err = c.commandImplementation.Exec(command, &bufOut, &bufErr, nil)
	if err != nil {
		return
	}

	stdOut = bufOut.String()
	stdErr = bufErr.String()

	return
}

func (c *Command) Run(parameters *[]byte) (output []byte, err error) {
	err = c.commandImplementation.Prepare(parameters)
	if err != nil {
		return
	}
	defer c.commandImplementation.CleanUp()

	for _, execSingle := range c.commandImplementation.Config().Execs {
		_, err = c.commandImplementation.Exec(execSingle, os.Stdout, os.Stderr, nil)
		if err != nil {
			return
		}
	}

	return c.commandImplementation.Output()
}
