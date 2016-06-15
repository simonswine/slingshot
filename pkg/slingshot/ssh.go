package slingshot

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"text/template"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"gopkg.in/yaml.v2"
)

type SshConfig struct {
	ConfigPath     string
	IdentityPath   string
	KnownHostsPath string
	Hosts          map[string][]ParameterInventory
}

func increment(i int) int {
	return i + 1
}

func (c *Cluster) sshConfig() string {
	inventoryString := `
- name: i-d45f3658
  publicIP: null
  privateIP: 172.20.129.5
  roles:
  - master
- name: i-bb5f3637
  publicIP: null
  privateIP: 172.20.129.208
  roles:
  - worker
- name: i-ba5f3636
  publicIP: null
  privateIP: 172.20.129.207
  roles:
  - worker
- name: i-0453c578670603810
  publicIP: 52.17.26.95
  privateIP: 172.20.3.97
  roles:
  - bastion`
	config := SshConfig{
		ConfigPath: path.Join(
			c.configDirPath(),
			"config/ssh-config",
		),
		KnownHostsPath: path.Join(
			c.configDirPath(),
			"config/ssh-known-hosts",
		),
		IdentityPath: path.Join(
			c.configDirPath(),
			"config/id_slingshot",
		),
		Hosts: make(map[string][]ParameterInventory),
	}

	inventory := []ParameterInventory{}
	err := yaml.Unmarshal([]byte(inventoryString), &inventory)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	for _, host := range inventory {
		for _, role := range host.Roles {
			if _, ok := config.Hosts[role]; !ok {
				config.Hosts[role] = []ParameterInventory{host}
			} else {
				config.Hosts[role] = append(config.Hosts[role], host)
			}
		}
	}
	funcMap := template.FuncMap{"increment": increment}

	tmpl, err := template.New("ssh-config").Funcs(funcMap).Parse(`{{$config := . -}}
{{range $role, $hosts := $config.Hosts -}}
{{range $hostIndex, $host := $hosts -}}
Host {{$host.Name}} {{$role}}{{if gt (len $hosts) 1}}{{increment $hostIndex}}{{end}}
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

{{end -}}
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
