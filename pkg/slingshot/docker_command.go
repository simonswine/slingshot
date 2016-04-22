package slingshot

import (
	"bytes"
	"errors"
	"io"
	"path"

	"github.com/fsouza/go-dockerclient"
	"github.com/simonswine/slingshot/pkg/utils"
)

var DockerSleepCommand = []string{"/bin/sleep", "3600"}
var DockerDefaultEntrypoint = []string{"/bin/sh", "-c"}

type DockerCommand struct {
	BaseCommand
	containerId *string
	entrypoint  *[]string
	workDir     *string
}

func (c *DockerCommand) getImageConfig() {
	inspect, err := c.provider.Docker().InspectImage(*c.provider.DockerImageId())
	if err != nil {
		c.log().Warn("failed to detect image config: ", err)
		c.entrypoint = &DockerDefaultEntrypoint
		defaultDir := "/"
		c.workDir = &defaultDir
	}
	c.workDir = &inspect.Config.WorkingDir
	c.entrypoint = &inspect.Config.Entrypoint
}

func (c *DockerCommand) Prepare(parameters *[]byte) error {
	c.getImageConfig()

	container, err := c.provider.Docker().CreateContainer(docker.CreateContainerOptions{
		Config: &docker.Config{
			Image:      *c.provider.DockerImageId(),
			Cmd:        DockerSleepCommand,
			Entrypoint: []string{},
		},
	})
	c.containerId = &container.ID
	if err != nil {
		return err
	}

	err = c.provider.Docker().StartContainer(*c.containerId, nil)
	if err != nil {
		return err
	}

	// write parameter file if needed
	if c.config != nil && c.config.ParameterFile != nil && parameters != nil {
		filePath := path.Join(
			*c.workDir,
			*c.config.ParameterFile,
		)
		err := c.uploadFile(
			filePath,
			*parameters,
			0644,
		)
		if err != nil {
			return err
		}
		c.log().Debugf("wrote parameters file to '%s'", filePath)
	}

	return nil
}

func (c *DockerCommand) ReadTar(statePaths []string) (tarData []byte, err error) {

	var tarArchives [][]byte

	for _, statePath := range statePaths {
		filePath := path.Join(
			*c.workDir,
			statePath,
		)
		buf := new(bytes.Buffer)
		err = c.provider.Docker().DownloadFromContainer(
			*c.containerId,
			docker.DownloadFromContainerOptions{
				Path:         filePath,
				OutputStream: buf,
			},
		)
		if err != nil {
			c.log().Debugf("skip storing state for %s : %s", statePath, err)
			continue
		}
		tarArchives = append(tarArchives, buf.Bytes())
	}

	if len(tarArchives) == 0 {
		return []byte{}, errors.New("No files to persist")
	}

	return utils.MergeTar(tarArchives)
}

func (c *DockerCommand) ExtractTar(tarData []byte, destPath string) error {
	tarReader := bytes.NewReader(tarData)
	containerPath := path.Join(
		*c.workDir,
		destPath,
	)
	if containerPath == "" {
		containerPath = "/"
	}
	return c.provider.Docker().UploadToContainer(
		*c.containerId,
		docker.UploadToContainerOptions{
			Path:        containerPath,
			InputStream: tarReader,
		},
	)
}

func (c *DockerCommand) uploadFile(filePath string, fileBody []byte, fileMode int64) (err error) {
	tarReader, err := utils.TarFromFile(
		path.Base(filePath),
		fileBody,
		fileMode,
	)
	if err != nil {
		return err
	}

	return c.provider.Docker().UploadToContainer(
		*c.containerId,
		docker.UploadToContainerOptions{
			Path:        path.Dir(filePath),
			InputStream: tarReader,
		},
	)
}

func (c *DockerCommand) downloadFile(path string) (content []byte, err error) {

	buf := new(bytes.Buffer)

	err = c.provider.Docker().DownloadFromContainer(
		*c.containerId,
		docker.DownloadFromContainerOptions{
			Path:         path,
			OutputStream: buf,
		},
	)
	if err != nil {
		return
	}

	content, _, err = utils.FirstFileFromTar(buf)
	return

}

func (c *DockerCommand) CleanUp() {
	if c.containerId != nil {
		err := c.provider.Docker().RemoveContainer(docker.RemoveContainerOptions{
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

	var execIncludingEntrypoint []string

	if c.entrypoint != nil {
		execIncludingEntrypoint = append(*c.entrypoint)
	}

	execIncludingEntrypoint = append(execIncludingEntrypoint, execCommand...)

	createOpts := docker.CreateExecOptions{
		Cmd:       execIncludingEntrypoint,
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

	execDocker, err := c.provider.Docker().CreateExec(createOpts)
	if err != nil {
		return
	}

	c.log().WithField("command", execIncludingEntrypoint).Debugf("run command")
	err = c.provider.Docker().StartExec(
		execDocker.ID,
		startOpts,
	)
	if err != nil {
		return
	}

	execInspect, err := c.provider.Docker().InspectExec(execDocker.ID)
	if err != nil {
		return
	}

	exitCode = execInspect.ExitCode
	return
}

func (c *DockerCommand) Output() (output []byte, err error) {
	if c.config != nil && c.config.ResultFile != nil {
		filePath := path.Join(
			*c.workDir,
			*c.config.ResultFile,
		)

		c.log().Debugf("Read output from file '%s'", filePath)
		return c.downloadFile(filePath)
	}
	return
}
