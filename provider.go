package main

import (
	"fmt"

	"bytes"
	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
)

type Provider struct {
	imageRepo string
	imageTag  string
	imageId   string
	docker    *docker.Client
	log       *log.Entry
	slingshot *Slingshot
}

func (p *Provider) initLog(name string) {
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

func (p *Provider) discover() error {

	createOptions := docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: p.imageId,
			Cmd:   []string{"discover"},
		},
	}

	container, err := p.docker.CreateContainer(createOptions)
	if err != nil {
		return err
	}

	// make sure container gets cleanup up in any case
	defer func() {
		err = p.docker.RemoveContainer(docker.RemoveContainerOptions{
			ID:    container.ID,
			Force: true,
		})
		if err != nil {
			p.log.Warnf("cleanup of container failed")
		}
	}()

	p.log = p.log.WithField("container_id", container.ID[0:len(container.ID)-52])

	err = p.docker.StartContainer(container.ID, &docker.HostConfig{})
	if err != nil {
		return err
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
		return err
	}

	exitCode, err := p.docker.WaitContainer(container.ID)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("Discover failed with exitcode=%d: %s", exitCode, bufErr.String())
	}

	p.log.Info("Out", bufOut.String())
	p.log.Info("Err", bufErr.String())

	return nil

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
