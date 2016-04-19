package main

import (
	"fmt"
	"io/ioutil"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/simonswine/slingshot/utils"
	"gopkg.in/yaml.v2"
)

type Cluster struct {
	Name                   string
	parameters             *Parameters
	infrastructureProvider *InfrastructureProvider
	configProvider         *ConfigProvider
	slingshot              *Slingshot
	providers              []string
}

func NewCluster(slingshot *Slingshot) *Cluster {
	c := &Cluster{
		slingshot: slingshot,
	}

	// register providers
	c.providers = []string{
		"infrastructure",
		"config",
	}

	return c
}

func (c *Cluster) newProvider(providerName string, imageName string) error {

	var provider *Provider

	if providerName == "infrastructure" {
		c.infrastructureProvider = &InfrastructureProvider{}
		provider = &c.infrastructureProvider.Provider

	} else if providerName == "config" {
		c.configProvider = &ConfigProvider{}
		provider = &c.configProvider.Provider
	} else {
		return fmt.Errorf("provider '%s' not found", providerName)
	}
	provider.slingshot = c.slingshot
	provider.init(providerName)
	return provider.initImage(imageName)
}

func (c *Cluster) log() *log.Entry {
	return log.WithFields(log.Fields{
		"cluster_name": c.Name,
		"context":      "slingshot",
	})
}

func (c *Cluster) createParameters(context *cli.Context) []error {
	paramsMain := &Parameters{}
	paramsMain.Defaults()

	sshKeyPath, err := utils.VagrantKeyPath()
	if err != nil {
		return []error{
			fmt.Errorf("Error while determining vagrant ssh key path: ", err),
		}
	}
	if context.IsSet("ssh-key") {
		sshKeyPath = context.String("ssh-key")
	}
	sshKey, err := ioutil.ReadFile(sshKeyPath)
	if err != nil {
		return []error{
			fmt.Errorf("Error while reading ssh key from '%s':  %s", sshKeyPath, err),
		}
	}

	sshKeyString := string(sshKey)
	paramsMain.General.Authentication.Ssh.PrivateKey = &sshKeyString
	errs := paramsMain.Validate()
	if len(errs) == 0 {
		c.parameters = paramsMain
	}
	return errs
}

func (c *Cluster) writeParameters() error {
	// TODO Don't forget to do 600
	c.log().Warnf("Implement me")
	return nil
}

func (c *Cluster) Create(context *cli.Context) []error {
	errs := c.createParameters(context)
	if len(errs) > 0 {
		return errs
	}

	for _, providerName := range c.providers {
		flagName := fmt.Sprintf("%s-provider", providerName)
		imageName := context.String(flagName)
		if len(imageName) == 0 {
			errs = append(errs, fmt.Errorf("No value for '--%s' provided", flagName))
			continue
		}

		err := c.newProvider(providerName, imageName)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errs
	}

	return c.Apply(context)

}

func (c *Cluster) Apply(context *cli.Context) []error {
	// run infrastructure apply
	paramsMainBytes, err := yaml.Marshal(c.parameters)
	if err != nil {
		return []error{
			fmt.Errorf("Error while writing parameters file: ", err),
		}
	}
	log.Debugf("params for infra:\n%s", paramsMainBytes)

	output, err := c.infrastructureProvider.RunCommand("apply", &paramsMainBytes)
	if err != nil {
		return []error{
			fmt.Errorf("Error while running infratstructure provider: ", err),
		}
	}

	// check and merge output from infrastructure apply
	c.parameters.Parse(string(output))
	errs := c.parameters.Validate()
	if len(errs) > 0 {
		return errs
	}

	// run config apply
	paramsMainBytes, err = yaml.Marshal(c.parameters)
	if err != nil {
		return []error{
			fmt.Errorf("Error while writing parameters file: ", err),
		}
	}
	log.Debugf("params after merge with output from infra:\n%s", paramsMainBytes)

	output, err = c.configProvider.RunCommand("apply", &paramsMainBytes)
	if err != nil {
		return []error{
			fmt.Errorf("Error while running infratstructure provider: ", err),
		}
	}

	return []error{}
}
