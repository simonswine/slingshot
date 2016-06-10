package slingshot

import (
	"fmt"

	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
)

type Parameters struct {
	Custom    map[string]string    `yaml:"custom,omitempty"`
	General   ParametersGeneral    `yaml:"general"`
	Inventory []ParameterInventory `yaml:"inventory"`
}

func (p *Parameters) Parse(content string) error {
	err := yaml.Unmarshal([]byte(content), p)
	return err
}

func (p *Parameters) Validate() (errs []error) {
	errs = append(errs, p.General.Validate()...)
	errs = append(errs, p.validateInventory()...)
	return
}

func (p *Parameters) validateInventory() (errs []error) {
	for _, elem := range p.Inventory {
		errs = append(errs, elem.Validate()...)
	}
	return
}

func (p *Parameters) Defaults() {
	p.General.Defaults()
}

type ParametersGeneral struct {
	Authentication ParametersAuthentication
	Cluster        ParametersCluster
}

func (pG *ParametersGeneral) Parse(content string) error {
	err := yaml.Unmarshal([]byte(content), pG)
	return err
}

func (pG *ParametersGeneral) Validate() (errs []error) {
	errs = append(errs, pG.Cluster.Validate()...)
	errs = append(errs, pG.Authentication.Validate()...)
	return
}

func (pG *ParametersGeneral) Defaults() {
	pG.Authentication.Defaults()
	pG.Cluster.Defaults()
}

type ParametersCluster struct {
	Kubernetes ParametersKubernetes
	Machines   map[string]ParameterMachine
}

func (pC *ParametersCluster) Defaults() {
	pC.Kubernetes.Defaults()

	masterMachines := ParameterMachine{}
	masterMachines.Defaults()
	masterRoles := []string{"master"}
	masterMachines.Roles = &masterRoles
	masterMachines.Count = 1

	workerMachines := ParameterMachine{}
	workerMachines.Defaults()
	workerRoles := []string{"worker"}
	workerMachines.Roles = &workerRoles
	workerMachines.Count = 2
	workerInstanceType := "t2.large"
	workerMachines.InstanceType = &workerInstanceType
	workerCores := 2
	workerMachines.Cores = &workerCores
	workerMemory := 1024
	workerMachines.Memory = &workerMemory

	pC.Machines = map[string]ParameterMachine{
		"master": masterMachines,
		"worker": workerMachines,
	}
}

func (pC *ParametersCluster) ValidateMachines() (errs []error) {
	for _, elem := range pC.Machines {
		errs = append(errs, elem.Validate()...)
	}
	return
}

func (pC *ParametersCluster) Validate() (errs []error) {
	errs = append(errs, pC.Kubernetes.Validate()...)
	errs = append(errs, pC.ValidateMachines()...)
	return
}

type ParameterMachine struct {
	Count        int       `yaml:"count"`
	Cores        *int      `yaml:"cores,omitempty"`
	Memory       *int      `yaml:"memory,omitempty"`
	InstanceType *string   `yaml:"instanceType,omitempty"`
	Roles        *[]string `yaml:"roles,omitempty"`
}

func (pM *ParameterMachine) Defaults() {
	pM.Count = 1
	cores := 1
	pM.Cores = &cores
	memory := 512
	pM.Memory = &memory
	roles := []string{"nodes"}
	pM.Roles = &roles
	instanceType := "m3.medium"
	pM.InstanceType = &instanceType
}

func (pM *ParameterMachine) Validate() (errs []error) {
	// TODO: Implement some validations
	return
}

type ParametersKubernetes struct {
	Interface      *string  `yaml:"interface,omitempty"`
	MasterApiPort  int      `yaml:"masterApiPort,omitempty"`
	MasterApiUrl   string   `yaml:"masterApiUrl,omitempty"`
	MasterSan      []string `yaml:"masterSan,omitempty"`
	ServiceNetwork string   `yaml:"serviceNetwork"`
	Dns            struct {
		Replicas   int
		DomainName string `yaml:"domainName"`
	}
	Networking string
	Flannel    struct {
		Subnet     string `yaml:"subnet"`
		Prefix     int
		HostPrefix int `yaml:"hostPrefix"`
	}
	Addons struct {
		ClusterLogging    bool `yaml:"clusterLogging"`
		ClusterMonitoring bool `yaml:"clusterMonitoring"`
		KubeUI            bool `yaml:"kubeUI"`
		KubeDash          bool `yaml:"kubeDash"`
	}
}

func (pK *ParametersKubernetes) Defaults() {
	pK.MasterApiPort = 443

	pK.ServiceNetwork = "10.245.0.0/16"

	pK.Dns.Replicas = 1
	pK.Dns.DomainName = "cluster.local"

	pK.Networking = "flannel"
	pK.Flannel.Subnet = "172.16.0.0"
	pK.Flannel.Prefix = 16
	pK.Flannel.HostPrefix = 24

	pK.Addons.ClusterLogging = false
	pK.Addons.ClusterMonitoring = false
	pK.Addons.KubeDash = false
	pK.Addons.KubeUI = false
}

func (pK *ParametersKubernetes) Validate() (errs []error) {
	errs = append(errs, pK.validateFlannel()...)
	return
}

func (pK *ParametersKubernetes) validateFlannel() (errs []error) {
	return
}

type ParametersAuthentication struct {
	Ssh struct {
		User       *string `yaml:"user,omitempty"`
		PrivateKey *string `yaml:"privateKey,omitempty"`
		PubKey     *string `yaml:"pubKey,omitempty"`
	}
}

func (pA *ParametersAuthentication) getPubKey() (pubKey string, err error) {
	key, err := ssh.ParsePrivateKey([]byte(*pA.Ssh.PrivateKey))
	if err != nil {
		return "", err
	}

	pubKey = string(ssh.MarshalAuthorizedKey(key.PublicKey()))

	// remove newline
	pubKey = pubKey[0 : len(pubKey)-1]
	return
}

func (pA *ParametersAuthentication) Validate() (errs []error) {
	if pA.Ssh.PrivateKey == nil {
		errs = append(errs, fmt.Errorf("Please provide a private key in general.authentication.privateKey"))
	} else if pA.Ssh.PubKey == nil {
		pubKey, err := pA.getPubKey()
		if err != nil {
			errs = append(errs, err)
		} else {
			pA.Ssh.PubKey = &pubKey
		}
	}
	return
}

func (pA *ParametersAuthentication) Defaults() {
	user := "root"
	pA.Ssh.User = &user
}

type ParameterInventory struct {
	Name      *string  `yaml:"name"`
	PublicIP  *string  `yaml:"publicIP"`
	PrivateIP *string  `yaml:"privateIP"`
	Roles     []string `yaml:"roles"`
}

func (pI *ParameterInventory) Validate() (errs []error) {
	if pI.PrivateIP == nil {
		errs = append(errs, fmt.Errorf("Required PrivateIP field missing"))
	}
	if len(pI.Roles) < 1 {
		errs = append(errs, fmt.Errorf("You need to specify at least one role"))
	}
	return
}
