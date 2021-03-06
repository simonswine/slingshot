package slingshot

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

type BaseCommand struct {
	config   *CommandConfig
	provider ProviderInterface
}

func (c *BaseCommand) Log() *log.Entry {
	return c.log()
}

func (c *BaseCommand) log() *log.Entry {
	commandType := "unknown"

	if c.config != nil {
		commandType = c.config.Type
	}

	l := log.WithFields(log.Fields{
		"context": fmt.Sprintf("%s-command", commandType),
	})

	return l
}

func (c *BaseCommand) Config() *CommandConfig {
	return c.config
}
