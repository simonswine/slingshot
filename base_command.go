package main

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
)

type BaseCommand struct {
	config *CommandConfig
}

func (c *BaseCommand) log() *log.Entry {
	l := log.WithFields(log.Fields{
		"context": fmt.Sprintf("%s-command", c.config.Type),
	})
	return l
}

func (c *BaseCommand) Config() *CommandConfig {
	return c.config
}
