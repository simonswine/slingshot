package main

import (
	"fmt"

	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"gopkg.in/yaml.v2"
)

type Provider struct {
	imageRepo    string
	imageTag     string
	imageId      string
	providerType string
	docker       *docker.Client
	log          *log.Entry
	slingshot    *Slingshot
	config       ProviderConfig
}

type ProviderConfig struct {
	Provider struct {
		Version string
		Type    string
	}
	Commands map[string]ProviderCommandConfig
}

type ProviderCommandConfig struct {
	Output     string `yaml:"output"`
	Type       string `yaml:"type"`
	PwdContent string `yaml:"pwdContent"`
}

func (c *ProviderConfig) Parse(content string) error {

	err := yaml.Unmarshal([]byte(content), c)

	return err
}

func (p *Provider) init(name string) {
	p.providerType = name
	p.log = log.WithFields(log.Fields{
		"context": name,
	})
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

func (p *Provider) getImage() (string, error) {

	dockerClient, err := p.slingshot.Docker()
	if err != nil {
		p.log.Error(err)
		return "", err
	}
	p.docker = dockerClient

	list, err := p.listImages()
	if err != nil {
		p.log.Error(err)
		return "", err
	}

	if len(list) == 0 {
		p.log.Debugf("pulling image from registry")

		err = p.pullImage()
		if err != nil {
			p.log.Error(err)
			return "", err
		}

		list, err = p.listImages()
		if err != nil {
			p.log.Error(err)
			return "", err
		}

	}

	if len(list) == 1 {
		return list[0].ID, nil

	} else {
		err = fmt.Errorf("This should never happen: more than a one image found (%d)", len(list))
		p.log.Error(err)
		return "", err
	}

	return "", nil
}

func (p *Provider) ImageName() string {
	return fmt.Sprintf("%s:%s", p.imageRepo, p.imageTag)
}

func (p *Provider) Execute(command []string) (stdOut string, stdErr string, exitCode int, err error) {

	container, err := p.docker.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: p.imageId,
			Cmd:   command,
		},
	})
	if err != nil {
		return
	}

	// make sure container gets cleanup up in any case
	defer func() {
		err = p.docker.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		if err != nil {
			p.log.Warnf("cleanup of container failed")
			return
		}
		p.log.Debugf("cleaned up container")
	}()

	// append container_id to log
	p.log = p.log.WithField("container_id", container.ID[0:len(container.ID)-52])

	err = p.docker.StartContainer(container.ID, &docker.HostConfig{})
	if err != nil {
		return
	}

	var bufOut bytes.Buffer
	var bufErr bytes.Buffer

	err = p.docker.AttachToContainer(docker.AttachToContainerOptions{
		Container:    container.ID,
		Stderr:       true,
		Stdout:       true,
		Stream:       true,
		Logs:         true,
		OutputStream: &bufOut,
		ErrorStream:  &bufErr,
	})
	if err != nil {
		return
	}

	exitCode, err = p.docker.WaitContainer(container.ID)
	if err != nil {
		return
	}

	err = nil
	stdErr = bufErr.String()
	stdOut = bufOut.String()

	return
}

func (p *Provider) discover() error {

	stdOut, stdErr, exitCode, err := p.Execute([]string{"discover"})
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("Discover failed with exitcode=%d: %s", exitCode, stdErr)
	}

	return p.config.Parse(stdOut)
}

func (p *Provider) initImage(imageName string) (err error) {
	// append latest tag if no tag
	p.imageRepo, p.imageTag = docker.ParseRepositoryTag(imageName)
	if len(p.imageTag) == 0 {
		p.imageTag = "latest"
	}
	// update logger
	p.log = p.log.WithField("image", p.ImageName())

	// get image
	p.imageId, err = p.getImage()
	if err != nil {
		p.log.Error(err)
		return err
	}

	p.log = p.log.WithField("image_id", p.imageId[0:len(p.imageId)-52])
	p.log.Info("found image")

	err = p.discover()
	if err != nil {
		p.log.Error(err)
		return err
	}

	return nil

}
