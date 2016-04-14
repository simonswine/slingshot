package main

import (
	"io"

	"github.com/fsouza/go-dockerclient"
)

type DockerCommand struct {
	BaseCommand
	dockerClient *docker.Client
	imageId      string
	containerId  *string
	stdIn        *io.PipeWriter
}

func (c *DockerCommand) Prepare() error {

	container, err := c.dockerClient.CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image: c.imageId,
			Cmd:   []string{"/bin/sleep", "3600"},
		},
	})
	c.containerId = &container.ID
	if err != nil {
		return err
	}

	return c.dockerClient.StartContainer(*c.containerId, nil)
}

func (c *DockerCommand) CleanUp() {
	if c.containerId != nil {
		err := c.dockerClient.RemoveContainer(docker.RemoveContainerOptions{
			ID:    *c.containerId,
			Force: true,
		})
		if err != nil {
			c.log().Warnf("cleanup of container failed")
		}
		c.containerId = nil
	}
}

func (c *DockerCommand) Exec(execCommand []string, stdout io.Writer, stderr io.Writer, stdin io.Reader) (exitCode int, err error) {
	createOpts := docker.CreateExecOptions{
		Cmd:       execCommand,
		Container: *c.containerId,
	}
	startOpts := docker.StartExecOptions{}

	if stdout != nil {
		startOpts.OutputStream = stdout
		createOpts.AttachStdout = true
	}
	if stderr != nil {
		startOpts.ErrorStream = stderr
		createOpts.AttachStderr = true
	}
	if stdin != nil {
		startOpts.InputStream = stdin
		createOpts.AttachStdin = true
	}

	execDocker, err := c.dockerClient.CreateExec(createOpts)
	if err != nil {
		return
	}

	err = c.dockerClient.StartExec(
		execDocker.ID,
		startOpts,
	)
	if err != nil {
		return
	}

	execInspect, err := c.dockerClient.InspectExec(execDocker.ID)
	if err != nil {
		return
	}

	exitCode = execInspect.ExitCode
	return
}
