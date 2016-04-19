package main

import (
	"fmt"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	"github.com/simonswine/slingshot/Godeps/_workspace/src/gopkg.in/yaml.v2"
	"github.com/simonswine/slingshot/utils"
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

func (s *Slingshot) readFile(path string) (string, error) {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(content), err
}

func (s *Slingshot) clusterCreateAction(context *cli.Context) {

	errs := []error{}

	paramsMain := &Parameters{}
	paramsMain.Defaults()

	sshKeyPath, err := utils.VagrantKeyPath()
	if err != nil {
		log.Fatal("Error while determining vagrant ssh key path: ", err)
	}
	if context.IsSet("ssh-key") {
		sshKeyPath = context.String("ssh-key")
	}
	sshKey, err := s.readFile(sshKeyPath)
	if err != nil {
		log.Errorf("Error while reading ssh key from '%s':  %s", sshKeyPath, err)
	}
	paramsMain.General.Authentication.Ssh.PrivateKey = &sshKey

	errs = paramsMain.Validate()

	for _, providerName := range s.providers {
		flagName := fmt.Sprintf("%s-provider", providerName)
		imageName := context.String(flagName)
		if len(imageName) == 0 {
			errs = append(errs, fmt.Errorf("No value for '--%s' provided", flagName))
			continue
		}

		err := s.newProvider(providerName, imageName)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		for _, err := range errs {
			log.Error(err)
		}
		log.Fatal("Errors prevent further execution")
	}

	// run infrastructure apply
	paramsMainBytes, err := yaml.Marshal(paramsMain)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("params for infra:\n%s", paramsMainBytes)

	output, err := s.infrastructureProvider.RunCommand("apply", &paramsMainBytes)
	if err != nil {
		log.Fatal(err)
	}

	// check and merge output from infrastructure apply
	paramsMain.Parse(string(output))
	errs = paramsMain.validateInventory()
	if len(errs) > 0 {
		for _, err := range errs {
			log.Error(err)
		}
		log.Fatal("Errors prevent further execution")
	}

	// run config apply
	paramsMainBytes, err = yaml.Marshal(paramsMain)
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("params after merge with output from infra:\n%s", paramsMainBytes)

	output, err = s.configProvider.RunCommand("apply", &paramsMainBytes)
	if err != nil {
		log.Fatal(err)
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
