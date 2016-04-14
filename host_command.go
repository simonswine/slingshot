package main

import (
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"syscall"

	"github.com/simonswine/slingshot/utils"
)

type HostCommand struct {
	BaseCommand
	oldWorkDir  *string
	tempWorkDir *string
}

func (c *HostCommand) Prepare() error {
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

	// skip untar if not existing
	if c.config == nil || len(c.config.WorkingDirContent) == 0 {
		return nil
	}

	return utils.UnTarGz([]byte(c.config.WorkingDirContent), *c.tempWorkDir)
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
