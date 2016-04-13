package main

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
)

type Slingshot struct {
	infrastructureProvider *InfrastructureProvider
	configProvider         *ConfigProvider
	log                    *log.Entry
	providers              []string
	dockerClient           *docker.Client
}

func NewSlingshot() *Slingshot {
	s := &Slingshot{}

	// init logger
	s.log = log.WithFields(log.Fields{
		"context": "slingshot",
	})

	// register providers
	s.providers = []string{
		"infrastructure",
		"config",
	}

	return s
}

func (s *Slingshot) Docker() (*docker.Client, error) {
	if s.dockerClient != nil {
		return s.dockerClient, nil
	}
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	env, err := client.Version()
	if err != nil {
		s.log.Fatal("connecting to docker failed: ", err)
		return nil, err
	}
	s.log.Debugf("connected to docker %+v", env.Get("Version"))

	s.dockerClient = client
	return s.dockerClient, nil
}

func (s *Slingshot) newProvider(providerName string, imageName string) error {

	var provider *Provider

	if providerName == "infrastructure" {
		s.infrastructureProvider = &InfrastructureProvider{}
		provider = &s.infrastructureProvider.Provider

	} else if providerName == "config" {
		s.configProvider = &ConfigProvider{}
		provider = &s.configProvider.Provider
	} else {
		return fmt.Errorf("provider '%s' not found", providerName)
	}
	provider.slingshot = s
	provider.init(providerName)
	return provider.initImage(imageName)
}

func (s *Slingshot) clusterCreateAction(context *cli.Context) {

	errs := []error{}

	for _, providerName := range s.providers {
		imageName := context.String(fmt.Sprintf("%s-provider", providerName))
		err := s.newProvider(providerName, imageName)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {

	}
}

func (s *Slingshot) unimplementedAction(context *cli.Context) {
	s.log.Warnf("Command '%s' (%s) not implemented", context.Command.HelpName, context.Command.Usage)
}

func (s *Slingshot) clusterCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "create",
			Usage:  "add a new cluster",
			Action: s.clusterCreateAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "infrastructure-provider, I",
					Usage: "Image name of the infrastructure provider to use",
				},
				cli.StringFlag{
					Name:  "config-provider, C",
					Usage: "Image name of the config provider to use",
				},
			},
		},
		{
			Name:  "list",
			Usage: "list existing clusters",
			// TODO Implement me
			Action: s.unimplementedAction,
		},
	}
}

func (s *Slingshot) Commands() []cli.Command {
	return []cli.Command{
		{
			Name:        "cluster",
			Usage:       "manage clusters",
			Subcommands: s.clusterCommands(),
		},
	}
}
