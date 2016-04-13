package main

import (
	"crypto/md5"
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestProviderConfigParseYamlInfrastructureHostCommand(t *testing.T) {

	yamlContent := `provider:
provider:
  type: infrastructure
  version: 1
commands:
  apply:
    type: hostCommand
    output: output.yml
    pwdContent: !!binary |
      H4sIAJJsDlcAA+3STQrCMBBA4Vl7ipygzUzT5DxFCi2IQhvB4/uH6EZQIXHzvs1sAjPw0rRSnL9I
      fX+dfUz+dT6IWvJdiNHMxKtq8OL68qeJHNc8LM7JdlrmNc/D/u27cVlrHFRX02rxHdfAMYRv+puq
      OF/8MqF/m8c1a5NPudiOz/uHEDVd+lvqOvrXcO9v/+2v4dnf7v2D0b+GadjtDu72Bzb/vgUAAAAA
      AAAAAAAAAPzmDI9fT9gAKAAA`

	c := &ProviderConfig{}
	err := c.Parse(yamlContent)

	assert.Nil(t, err, "Unexpected error during parsing")

	assert.Equal(t, "1", c.Provider.Version)
	assert.Equal(t, "infrastructure", c.Provider.Type)

	// testing first command
	assert.Equal(t, "hostCommand", c.Commands["apply"].Type)
	assert.Equal(t, "output.yml", c.Commands["apply"].Output)

	assert.Equal(
		t,
		"716248193fe94b19ce40865106938f8b",
		fmt.Sprintf("%x", md5.Sum([]byte(c.Commands["apply"].PwdContent))),
	)
}
