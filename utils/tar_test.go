package utils_test

import (
	"testing"
	"io/ioutil"

	"github.com/simonswine/slingshot/utils"
	"github.com/stretchr/testify/assert"
)


func  TestTarFromFile(t *testing.T) {

	tarReader, err := utils.TarFromFile(
		"test.txt",
		[]byte("test123"),
		0644,
	)
	assert.Nil(t, err, "Error during tar building")

	tarBytes, err := ioutil.ReadAll(tarReader)
	assert.Nil(t, err, "Error during tar reading")

	assert.True(t, 50 < len(tarBytes))
}
