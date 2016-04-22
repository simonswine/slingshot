package slingshot

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
	"path"
)

type Provider struct {
	imageRepo    string
	imageTag     string
	imageId      *string
	containerId  *string
	providerType string
	docker       *docker.Client
	cluster      *Cluster
	config       ProviderConfig
}

type ProviderInterface interface {
	StatePath() string
	Log() *log.Entry
	Docker() *docker.Client
	DockerImageId() *string
}

type ProviderConfig struct {
	Provider struct {
		Version string
		Type    string
	}
	Commands map[string]CommandConfig
}

func (c *ProviderConfig) Parse(content string) error {

	err := yaml.Unmarshal([]byte(content), c)

	return err
}

func (p *Provider) init(name string) {
	p.providerType = name

}

func (p *Provider) Log() *log.Entry {

	l := log.WithFields(log.Fields{
		"context": fmt.Sprintf("%s-provider", p.providerType),
	})

	l = l.WithField("image", p.ImageName())

	if p.imageId != nil {
		s := *p.imageId
		l = l.WithField("image_id", s[0:len(s)-52])
	}

	if p.containerId != nil {
		s := *p.containerId
		l = l.WithField("container_id", s[0:len(s)-52])
	}

	return l
}

func (p *Provider) RunCommand(commandName string, parameters *[]byte) (output []byte, err error) {
	p.Log().Debugf("running command '%s'", commandName)

	if commandDef, ok := p.config.Commands[commandName]; ok {
		c, errCmd := NewCommand(&commandDef, p)
		if errCmd != nil {
			err = errCmd
			return
		}
		return c.Run(parameters)

	}

	err = fmt.Errorf("command '%s' not found", commandName)
	return
}

func (p *Provider) pullImage() error {
	// TODO: Support private reg auth
	//auth, err := docker.NewAuthConfigurationsFromDockerCfg()
	//if err != nil {
	//	return err
	//}

	return p.docker.PullImage(
		docker.PullImageOptions{
			Repository: p.imageRepo,
			Tag:        p.imageTag,
		},
		docker.AuthConfiguration{},
	)
}

func (p *Provider) listImages() ([]docker.APIImages, error) {
	return p.docker.ListImages(docker.ListImagesOptions{
		All:    false,
		Filter: p.ImageName(),
	})
}

func (p *Provider) Docker() *docker.Client {
	dockerClient, _ := p.cluster.slingshot.Docker()
	return dockerClient
}

func (p *Provider) DockerImageId() *string {
	return p.imageId
}

func (p *Provider) StatePath() string {
	return path.Join(
		p.cluster.configDirPath(),
		fmt.Sprintf("provider-%s.tar", p.providerType),
	)
}

func (p *Provider) getImage() (string, error) {

	dockerClient, err := p.cluster.slingshot.Docker()
	if err != nil {
		p.Log().Error(err)
		return "", err
	}
	p.docker = dockerClient

	list, err := p.listImages()
	if err != nil {
		p.Log().Error(err)
		return "", err
	}

	if len(list) == 0 {
		p.Log().Debugf("pulling image from registry")

		err = p.pullImage()
		if err != nil {
			p.Log().Error(err)
			return "", err
		}

		list, err = p.listImages()
		if err != nil {
			p.Log().Error(err)
			return "", err
		}

	}

	if len(list) == 1 {
		return list[0].ID, nil

	}

	err = fmt.Errorf("This should never happen: more than a one image found (%d)", len(list))
	p.Log().Error(err)
	return "", err
}

func (p *Provider) ImageName() string {
	return fmt.Sprintf("%s:%s", p.imageRepo, p.imageTag)
}

func (p *Provider) readConfig() error {

	c, err := NewCommand(
		&CommandConfig{
			Type: "docker",
		},
		p,
	)
	if err != nil {
		return err
	}

	stdOut, stdErr, exitCode, err := c.Execute([]string{"discover"})
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("discover failed with exitcode=%d: %s", exitCode, stdErr)
	}

	return p.config.Parse(stdOut)
}

func (p *Provider) initImage(imageName string) (err error) {
	// append latest tag if no tag
	p.imageRepo, p.imageTag = docker.ParseRepositoryTag(imageName)
	if len(p.imageTag) == 0 {
		p.imageTag = "latest"
	}

	// get image
	imageId, err := p.getImage()
	if err != nil {
		p.Log().Error(err)
		return err
	}
	p.imageId = &imageId

	p.Log().Info("found image")

	err = p.readConfig()
	if err != nil {
		p.Log().Error(err)
		return err
	}

	return nil

}
