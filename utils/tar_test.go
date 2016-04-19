package utils

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"os"
	"path"
)

func createExampleTree(tempDir string) error {
	err := os.Mkdir(
		path.Join(tempDir, "testdir"),
		0750,
	)
	if err != nil {
		return err
	}

	err = os.Mkdir(
		path.Join(tempDir, "testempty"),
		0750,
	)
	if err != nil {
		return err
	}

	ioutil.WriteFile(
		path.Join(tempDir, "test1.txt"),
		[]byte("test1"),
		0600,
	)
	if err != nil {
		return err
	}

	ioutil.WriteFile(
		path.Join(tempDir, "testdir", "test2.txt"),
		[]byte("test2"),
		0640,
	)
	if err != nil {
		return err
	}
	return nil
}

func TestTarFromFile(t *testing.T) {

	tarReader, err := TarFromFile(
		"test.txt",
		[]byte("test123"),
		0644,
	)
	assert.Nil(t, err, "Error during tar building")

	tarBytes, err := ioutil.ReadAll(tarReader)
	assert.Nil(t, err, "Error during tar reading")

	assert.True(t, 50 < len(tarBytes))
}

func TestWalkDirFromObject(t *testing.T) {

	tempDir, err := ioutil.TempDir("", "gotest")
	if err != nil {
		t.Error(err)
	}

	err = createExampleTree(tempDir)
	if err != nil {
		t.Error(err)
	}

	objects, err := walkDirToObjects(tempDir, tempDir)

	assert.Nil(t, err, "No error happens")
	assert.Equal(t, 4, len(objects))

	testFiles := 0
	for _, object := range objects {
		if object.Header.Name == "test1.txt" {
			assert.Equal(t, int64(0600), object.Header.Mode)
			assert.Equal(t, "test1", string(*object.Body))
			testFiles++
		}
		if object.Header.Name == "testdir/test2.txt" {
			assert.Equal(t, int64(0640), object.Header.Mode)
			assert.Equal(t, "test2", string(*object.Body))
			testFiles++
		}
	}
	assert.Equal(t, 2, testFiles, "not engough files found")
}
