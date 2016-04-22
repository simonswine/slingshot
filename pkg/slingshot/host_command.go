package slingshot

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"syscall"

	"github.com/simonswine/slingshot/pkg/utils"
)

type HostCommand struct {
	BaseCommand
	oldWorkDir  *string
	tempWorkDir *string
}

func (c *HostCommand) Prepare(parameters *[]byte) error {
	oldWorkDir, err := os.Getwd()
	if err != nil {
		return err
	}
	c.oldWorkDir = &oldWorkDir

	tempWorkDir, err := ioutil.TempDir("", AppName)
	if err != nil {
		return err
	}
	c.tempWorkDir = &tempWorkDir

	err = os.Chdir(*c.tempWorkDir)
	if err != nil {
		return err
	}

	// untar work dir if needed
	if c.config != nil && len(c.config.WorkingDirContent) != 0 {
		err = utils.UnTarGz([]byte(c.config.WorkingDirContent), *c.tempWorkDir)
		if err != nil {
			return err
		}
	}

	// write parameter file if needed
	if c.config != nil && c.config.ParameterFile != nil && parameters != nil {
		filePath := path.Join(
			*c.tempWorkDir,
			*c.config.ParameterFile,
		)
		err = ioutil.WriteFile(
			filePath,
			[]byte(*parameters),
			0644,
		)
		if err != nil {
			return err
		}
		c.log().Debugf("wrote parameters file to '%s'", filePath)
	}

	return nil
}

func (c *HostCommand) CleanUp() {
	if c.oldWorkDir != nil {
		err := os.Chdir(*c.oldWorkDir)
		if err != nil {
			c.log().Warn(err)
		}
		c.oldWorkDir = nil
	}

	if c.tempWorkDir != nil {
		err := os.RemoveAll(*c.tempWorkDir)
		if err != nil {
			c.log().Warn(err)
		}
		c.tempWorkDir = nil
	}
}

func (c *HostCommand) Exec(execSingle []string, stdout io.Writer, stderr io.Writer, stdin io.Reader) (exitCode int, err error) {
	cmd := exec.Command(execSingle[0], execSingle[1:len(execSingle)]...)
	if stdout != nil {
		cmd.Stdout = stdout

	}
	if stderr != nil {
		cmd.Stderr = stderr
	}
	if stdin != nil {
		cmd.Stdin = stdin
	}

	err = cmd.Start()
	if err != nil {
		return
	}

	err = cmd.Wait()
	if err != nil {

		// detect right exitCode
		if exiterr, ok := err.(*exec.ExitError); ok {
			//err = nil
			if status, ok := exiterr.Sys().(syscall.WaitStatus); ok {
				err = nil
				exitCode = status.ExitStatus()
			}
		}
	}

	return
}

func (c *HostCommand) ReadTar(statePaths []string) (tarData []byte, err error) {

	var tarObjects []utils.TarObject

	for _, statePath := range statePaths {
		objs, err := utils.WalkDirToObjects(
			path.Join(
				*c.tempWorkDir,
				statePath,
			),
			*c.tempWorkDir,
		)
		if err != nil {
			return []byte{}, err
		}
		tarObjects = append(tarObjects, objs...)
	}

	return utils.TarListOfObjects(tarObjects)
}

func (c *HostCommand) ExtractTar(tarData []byte, destPath string) error {
	err := utils.UnTar(
		tarData,
		*c.tempWorkDir,
	)
	return err
}

func (c *HostCommand) Output() (output []byte, err error) {
	if c.config != nil && c.config.ResultFile != nil && c.tempWorkDir != nil {
		filePath := path.Join(
			*c.tempWorkDir,
			*c.config.ResultFile,
		)
		c.log().Debugf("Read output from file '%s'", filePath)
		return ioutil.ReadFile(filePath)
	}
	return
}
