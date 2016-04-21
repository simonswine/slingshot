package main

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
	"path"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/fsouza/go-dockerclient"
	"github.com/stretchr/testify/assert"
)

type MockProvider struct {
	tmpDir *string
}

func (p *MockProvider) StatePath() string {
	if p.tmpDir == nil {
		tmpDir, _ := ioutil.TempDir("", AppName)
		p.tmpDir = &tmpDir
	}

	return path.Join(
		*p.tmpDir,
		"provider-test.tar",
	)
}

func (p *MockProvider) Log() *log.Entry {
	log.SetLevel(log.DebugLevel)
	return log.WithField("context", "mock-provider")
}

func (p *MockProvider) Docker() *docker.Client {
	dockerClient, _ := docker.NewClientFromEnv()
	return dockerClient
}

func (p *MockProvider) DockerImageId() *string {
	str := "busybox:latest"
	return &str
}

func TestProviderConfigParseYamlInfrastructureHostCommand(t *testing.T) {

	yamlContent := `provider:
provider:
  type: infrastructure
  version: 1
commands:
  apply:
    execs:
      -
        - vagrant
        - up
    type: host
    parameterFile: params.yaml
    resultFile: output.yaml
    persistPaths:
      - .vagrant/
    workingDirContent: !!binary |
      H4sIAJJsDlcAA+3STQrCMBBA4Vl7ipygzUzT5DxFCi2IQhvB4/uH6EZQIXHzvs1sAjPw0rRSnL9I
      fX+dfUz+dT6IWvJdiNHMxKtq8OL68qeJHNc8LM7JdlrmNc/D/u27cVlrHFRX02rxHdfAMYRv+puq
      OF/8MqF/m8c1a5NPudiOz/uHEDVd+lvqOvrXcO9v/+2v4dnf7v2D0b+GadjtDu72Bzb/vgUAAAAA
      AAAAAAAAAPzmDI9fT9gAKAAA`

	c := &ProviderConfig{}
	err := c.Parse(yamlContent)

	assert.Nil(t, err, "Unexpected error during parsing")

	// ensure all vars are parsed
	assert.Equal(t, "infrastructure", c.Provider.Type)
	assert.Equal(t, "1", c.Provider.Version)

	assert.Equal(t, [][]string{[]string{"vagrant", "up"}}, c.Commands["apply"].Execs)
	assert.Equal(t, "host", c.Commands["apply"].Type)
	assert.Equal(t, "params.yaml", *c.Commands["apply"].ParameterFile)
	assert.Equal(t, "output.yaml", *c.Commands["apply"].ResultFile)
	assert.Equal(t, []string{".vagrant/"}, c.Commands["apply"].PersistPaths)
	assert.Equal(
		t,
		"716248193fe94b19ce40865106938f8b",
		fmt.Sprintf("%x", md5.Sum([]byte(c.Commands["apply"].WorkingDirContent))),
	)
}
