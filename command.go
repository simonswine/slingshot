package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"

	log "github.com/Sirupsen/logrus"
)

type CommandInterface interface {
	Exec(command []string, stdout io.Writer, stderr io.Writer, stdin io.Reader) (exitCode int, err error)
	Prepare(*[]byte) error
	Output() ([]byte, error)
	CleanUp()
	Config() *CommandConfig
	Log() *log.Entry
	ReadTar([]string) ([]byte, error)
	ExtractTar([]byte, string) error
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
	provider              ProviderInterface
}

func NewCommand(c *CommandConfig, p *Provider) (*Command, error) {
	cmd := &Command{}
	err := cmd.Init(c, p)
	return cmd, err
}

func (c *Command) log() *log.Entry {
	return c.provider.Log()
}

func (c *Command) Init(config *CommandConfig, p ProviderInterface) error {

	c.provider = p

	if config.Type == "host" {
		c.commandImplementation = &HostCommand{
			BaseCommand: BaseCommand{
				config:   config,
				provider: p,
			},
		}
	} else if config.Type == "docker" {
		c.commandImplementation = &DockerCommand{
			BaseCommand: BaseCommand{
				config:   config,
				provider: p,
			},
		}
	} else {
		return fmt.Errorf("command type '%s' not found", config.Type)
	}

	return nil
}

func (c *Command) Execute(command []string) (stdOut string, stdErr string, exitCode int, err error) {
	err = c.Prepare(nil)
	if err != nil {
		return
	}
	defer c.CleanUp()

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
	err = c.Prepare(parameters)
	if err != nil {
		return
	}
	defer c.CleanUp()

	for _, execSingle := range c.commandImplementation.Config().Execs {
		_, err = c.commandImplementation.Exec(execSingle, os.Stdout, os.Stderr, nil)
		if err != nil {
			return
		}
	}

	return c.commandImplementation.Output()
}

func (c *Command) Prepare(parameters *[]byte) error {
	if err := c.commandImplementation.Prepare(parameters); err != nil {
		return err
	}

	// restore persisted state if needed
	conf := c.commandImplementation.Config()
	if conf != nil && len(conf.PersistPaths) > 0 {
		if err := c.restoreState(); err != nil {
			c.log().Warn("restoring of state failed: ", err)
		}
	}

	return nil
}

func (c *Command) CleanUp() {
	// persist state if needed
	conf := c.commandImplementation.Config()
	if conf != nil && len(conf.PersistPaths) > 0 {
		if err := c.persistState(conf.PersistPaths); err != nil {
			c.log().Warn("persisting of state failed: ", err)
		}
	}

	c.commandImplementation.CleanUp()
}

func (c *Command) persistState(paths []string) error {
	tarBytes, err := c.commandImplementation.ReadTar(paths)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(
		c.provider.StatePath(),
		tarBytes,
		0600,
	)
	if err != nil {
		return err
	}

	c.log().Debugf(
		"successfully stored state in file %s (%d bytes)",
		c.provider.StatePath(),
		len(tarBytes),
	)
	return nil
}

func (c *Command) restoreState() error {

	if _, err := os.Stat(c.provider.StatePath()); os.IsNotExist(err) {
		return nil
	}

	tarData, err := ioutil.ReadFile(c.provider.StatePath())
	if err != nil {
		return err
	}

	err = c.commandImplementation.ExtractTar(
		tarData,
		"",
	)
	if err != nil {
		return err
	}

	c.log().Debugf(
		"successfully restored state from file %s",
		c.provider.StatePath(),
	)
	return nil
}
