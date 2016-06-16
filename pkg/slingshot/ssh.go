package slingshot

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

type SshConfig struct {
	ConfigPath     string
	IdentityPath   string
	KnownHostsPath string
	Inventory      []ParameterInventory
}

func (c *Cluster) sshConfig() string {
	inventory, err := c.Inventory()
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	config := SshConfig{
		ConfigPath: path.Join(
			c.configDirPath(),
			SlingshotSSHConfigFileName,
		),
		KnownHostsPath: path.Join(
			c.configDirPath(),
			SlingshotSSHKnownHostsFileName,
		),
		IdentityPath: path.Join(
			c.configDirPath(),
			SlingshotSSHIdentityFileName,
		),
		Inventory: inventory,
	}

	tmpl, err := template.New("ssh-config").Parse(`{{$config := . -}}
{{range $hostIndex, $host := $config.Inventory -}}
Host {{$host.Name}}{{range $host.Aliases}} {{.}}{{end}}
{{- if $host.PublicIP}}
    Hostname {{$host.PublicIP}}
{{- else}}
    ProxyCommand ssh -F {{$config.ConfigPath}} -q bastion ncat %h 22
    Hostname {{$host.PrivateIP}}
{{- end}}
    IdentitiesOnly yes
    IdentityFile {{$config.IdentityPath}}
    UserKnownHostsFile {{$config.KnownHostsPath}}
    User core
    StrictHostKeyChecking no

{{end -}}`)
	if err != nil {
		log.Fatal(err)
	}
	f, err := os.OpenFile(config.ConfigPath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatal("error opening ssh-config file: ", err)
	}
	defer f.Close()

	err = tmpl.Execute(f, config)
	if err != nil {
		log.Fatal("error writing ssh-config file: ", err)
	}

	// write ssh id out
	err = ioutil.WriteFile(
		config.IdentityPath,
		[]byte(*c.Parameters.General.Authentication.Ssh.PrivateKey),
		0600,
	)
	if err != nil {
		log.Fatal("error writing ssh-id file: ", err)
	}

	return config.ConfigPath
}

func (c *Cluster) Ssh(context *cli.Context) {
	c.sshConfig()

	binary, err := exec.LookPath("ssh")
	if err != nil {
		log.Fatal(err)
	}

	args := []string{
		"-F",
		c.sshConfig(),
	}
	args = append(args, context.Args()[1:]...)

	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		log.Fatal(err)
	}

	err = cmd.Wait()
	if err != nil {
		log.Fatal(err)
	}
}
