package main

import (
	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
)

type Slingshot struct {
	log          *log.Entry
	dockerClient *docker.Client
	clusters     []*Cluster
}

func NewSlingshot() *Slingshot {
	s := &Slingshot{}

	// init logger
	s.log = log.WithFields(log.Fields{
		"context": "slingshot",
	})

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

func (s *Slingshot) clusterCreateAction(context *cli.Context) {
	c := NewCluster(s)

	s.clusters = append(s.clusters, c)

	errs := c.Create(context)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Error(err)
		}
		log.Fatal("Errors prevent further execution")
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
				cli.StringFlag{
					Name:  "ssh-key, i",
					Usage: "SSH private key to use (please provide an uncrypted key, default: vagrant insecure key)",
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
