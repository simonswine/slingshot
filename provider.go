package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/simonswine/slingshot/utils"
	"gopkg.in/yaml.v2"
)

type Provider struct {
	imageRepo    string
	imageTag     string
	imageId      *string
	containerId  *string
	providerType string
	docker       *docker.Client
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

}

func (p *Provider) log() *log.Entry {

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

func (p *Provider) RunCommand(commandName string) error {
	p.log().Debugf("running command '%s'", commandName)

	if commandDef, ok := p.config.Commands[commandName]; ok {
		if commandDef.Type == "hostCommand" {
			return p.runHostCommand(commandDef)
		} else {
			return fmt.Errorf("command type '%s' not found", commandDef.Type)
		}

	} else {
		return fmt.Errorf("command '%s' not found", commandName)
	}

	return nil
}

func (p *Provider) runHostCommand(def ProviderCommandConfig) error {

	oldWd, err := os.Getwd()
	if err != nil {
		return err
	}

	dir, err := ioutil.TempDir("", AppName)
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	err = os.Chdir(dir)
	if err != nil {
		return err
	}
	defer os.Chdir(oldWd)

	err = utils.UnTarGz([]byte(def.PwdContent), dir)

	cmd := exec.Command("vagrant", "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err != nil {
		return err
	}

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return err
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
		p.log().Error(err)
		return "", err
	}
	p.docker = dockerClient

	list, err := p.listImages()
	if err != nil {
		p.log().Error(err)
		return "", err
	}

	if len(list) == 0 {
		p.log().Debugf("pulling image from registry")

		err = p.pullImage()
		if err != nil {
			p.log().Error(err)
			return "", err
		}

		list, err = p.listImages()
		if err != nil {
			p.log().Error(err)
			return "", err
		}

	}

	if len(list) == 1 {
		return list[0].ID, nil

	} else {
		err = fmt.Errorf("This should never happen: more than a one image found (%d)", len(list))
		p.log().Error(err)
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
			Image: *p.imageId,
			Cmd:   command,
		},
	})
	p.containerId = &container.ID

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
			p.log().Warnf("cleanup of container failed")
			return
		}
		p.log().Debugf("cleaned up container")
		p.containerId = nil
	}()

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

func (p *Provider) readConfig() error {

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

	// get image
	imageId, err := p.getImage()
	if err != nil {
		p.log().Error(err)
		return err
	}
	p.imageId = &imageId

	p.log().Info("found image")

	err = p.readConfig()
	if err != nil {
		p.log().Error(err)
		return err
	}

	return nil

}
