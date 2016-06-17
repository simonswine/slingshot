package slingshot

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"regexp"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/simonswine/slingshot/pkg/utils"
	"gopkg.in/yaml.v2"
)

type Cluster struct {
	Version                string
	Parameters             *Parameters        `yaml:"parameters"`
	ProviderImageNames     map[string]*string `yaml:"providerImageNames"`
	infrastructureProvider *InfrastructureProvider
	configProvider         *ConfigProvider
	slingshot              *Slingshot
}

func NewCluster(slingshot *Slingshot) *Cluster {
	c := &Cluster{
		slingshot: slingshot,
		Version:   "1",
	}

	// initialize map
	c.ProviderImageNames = map[string]*string{
		"infrastructure": nil,
		"config":         nil,
	}

	return c
}

func LoadClusterFromPath(slingshot *Slingshot, filePath string) (*Cluster, error) {

	yamlData, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	c := NewCluster(slingshot)

	if err = yaml.Unmarshal(yamlData, c); err != nil {
		return nil, err
	}

	return c, nil
}
func (c *Cluster) initProviders() (errs []error) {
	for providerName, imageName := range c.ProviderImageNames {
		if imageName == nil {
			c.log().Warnf("Provider %s has no image name specified", providerName)
			continue
		}
		err := c.newProvider(providerName, *imageName)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return
}

func (c *Cluster) newProvider(providerName string, imageName string) error {

	var provider *Provider

	if providerName == "infrastructure" {
		c.infrastructureProvider = &InfrastructureProvider{}
		provider = &c.infrastructureProvider.Provider

	} else if providerName == "config" {
		c.configProvider = &ConfigProvider{}
		provider = &c.configProvider.Provider
	} else {
		return fmt.Errorf("provider '%s' not found", providerName)
	}
	provider.cluster = c
	provider.init(providerName)
	return provider.initImage(imageName)
}

func (c *Cluster) Validate() (errs []error) {
	errs = append(errs, c.validateName()...)
	return
}

func (c *Cluster) Name() string {
	return c.Parameters.General.Cluster.Name
}

func (c *Cluster) validateName() (errs []error) {

	if existingC, err := c.slingshot.getClusterByName(c.Name()); err == nil && existingC != c {
		return []error{fmt.Errorf("cluster with the name '%s' already exists", c.Name())}
	}

	if len(c.Name()) == 0 {
		return []error{fmt.Errorf("empty cluster name not allowed")}
	}

	regExpName := "[a-z0-9-]+"
	matched, err := regexp.MatchString(regExpName, c.Name())
	if err != nil {
		return []error{err}
	}
	if !matched {
		return []error{fmt.Errorf("cluster name '%s' did not match '%s'", c.Name(), regExpName)}
	}
	return []error{}
}

func (c *Cluster) configDirPath() string {
	return path.Join(
		c.slingshot.configDir,
		c.Name(),
	)
}

func (c *Cluster) configFilePath() string {
	return path.Join(
		c.configDirPath(),
		SlingshotClusterFileName,
	)
}

func (c *Cluster) inventoryFilePath() string {
	return path.Join(
		c.configDirPath(),
		SlingshotInventoryFileName,
	)
}

func (c *Cluster) kubectlFilePath() string {
	return path.Join(
		c.configDirPath(),
		SlingshotKubectlConfigFileName,
	)
}

func (c *Cluster) WriteConfig() error {
	return c.writeYaml(
		c.configFilePath(),
		c,
	)
}

func (c *Cluster) WriteInventory() error {
	return c.writeYaml(
		c.inventoryFilePath(),
		c.Parameters.Inventory,
	)
}

func (c *Cluster) Inventory() (inventory []ParameterInventory, err error) {

	reader, err := os.Open(c.inventoryFilePath())
	if err != nil {
		return
	}

	yamlData, err := ioutil.ReadAll(reader)
	if err != nil {
		return
	}

	err = yaml.Unmarshal(yamlData, &inventory)
	return
}

func (c *Cluster) writeFile(filePath string, body []byte) error {
	if err := utils.EnsureDirectory(path.Dir(filePath)); err != nil {
		return err
	}

	err := ioutil.WriteFile(filePath, body, 0600)
	if err != nil {
		return err
	}

	c.log().Infof("wrote to '%s'", filePath)
	return nil
}

func (c *Cluster) writeYaml(path string, obj interface{}) error {
	yamlContents, err := yaml.Marshal(obj)
	if err != nil {
		return err
	}

	return c.writeFile(path, yamlContents)
}

func (c *Cluster) log() *log.Entry {
	return log.WithFields(log.Fields{
		"cluster_name": c.Name,
		"context":      "slingshot",
	})
}

func (c *Cluster) createParameters(context *cli.Context) []error {
	paramsMain := &Parameters{}
	paramsMain.Defaults()

	// read cluster config if specified
	clusterFile := context.String("cluster-file")
	if len(clusterFile) > 0 {
		var reader io.Reader
		var err error
		if clusterFile == "-" {
			reader = os.Stdin
		} else {
			reader, err = os.Open(clusterFile)
			if err != nil {
				return []error{err}
			}
		}
		yamlData, err := ioutil.ReadAll(reader)
		if err != nil {
			return []error{fmt.Errorf("Error reading cluster-file: %s", err)}
		}

		if err = yaml.Unmarshal(yamlData, paramsMain); err != nil {
			return []error{fmt.Errorf("Error reading cluster-file: %s", err)}
		}
	}

	key, err := rsa.GenerateKey(
		rand.Reader,
		2048,
	)
	if err != nil {
		return []error{fmt.Errorf("Error generating ssh-key: %s", err)}
	}

	pemKey := pem.Block{
		Type:    "RSA PRIVATE KEY",
		Headers: nil,
		Bytes:   x509.MarshalPKCS1PrivateKey(key),
	}
	stringKey := string(pem.EncodeToMemory(&pemKey))

	paramsMain.General.Authentication.Ssh.PrivateKey = &stringKey
	errs := paramsMain.Validate()
	if len(errs) == 0 {
		c.Parameters = paramsMain
	}
	return errs
}

func (c *Cluster) Create(context *cli.Context) (errs []error) {

	// read cluster name
	clusterName, err := c.slingshot.readClusterName(context)
	if err != nil {
		errs = append(errs, err)
	}

	errs = append(errs, c.createParameters(context)...)
	if len(errs) > 0 {
		return errs
	}

	c.Parameters.General.Cluster.Name = clusterName
	errs = append(errs, c.validateName()...)

	// read provider flags
	for providerName, _ := range c.ProviderImageNames {
		flagName := fmt.Sprintf("%s-provider", providerName)
		imageName := context.String(flagName)
		c.ProviderImageNames[providerName] = &imageName
		if len(imageName) == 0 {
			errs = append(errs, fmt.Errorf("No value for '--%s' provided", flagName))
			continue
		}
	}
	if len(errs) > 0 {
		return errs
	}

	// write config
	err = c.WriteConfig()
	if err != nil {
		errs = append(errs, err)
	}

	return c.apply()

}

func (c *Cluster) Destroy(context *cli.Context) (errs []error) {
	return c.destroy()
}

func (c *Cluster) destroy() (errs []error) {

	errs = append(errs, c.initProviders()...)
	if len(errs) > 0 {
		return errs
	}

	// run infrastructure destroy
	paramsMainBytes, err := yaml.Marshal(c.Parameters)
	if err != nil {
		return []error{
			fmt.Errorf("Error while writing parameters file: %s", err),
		}
	}
	log.Debugf("params for infra:\n%s", paramsMainBytes)

	_, err = c.infrastructureProvider.RunCommand("destroy", &paramsMainBytes)
	if err != nil {
		return []error{
			fmt.Errorf("Error while running infrastructure provider: %s", err),
		}
	}

	return
}

func (c *Cluster) Kubectl(context *cli.Context) {
	binary, err := exec.LookPath("kubectl")
	if err != nil {
		log.Fatal(err)
	}

	cmd := exec.Command(binary, context.Args()[1:]...)
	env := os.Environ()
	env = append(env, fmt.Sprintf("KUBECONFIG=%s", c.kubectlFilePath()))
	cmd.Env = env
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

func (c *Cluster) Apply(context *cli.Context) (errs []error) {
	return c.apply()
}

func (c *Cluster) apply() (errs []error) {

	errs = append(errs, c.initProviders()...)
	if len(errs) > 0 {
		return errs
	}

	// run infrastructure apply
	paramsMainBytes, err := yaml.Marshal(c.Parameters)
	if err != nil {
		return []error{
			fmt.Errorf("Error while writing parameters file: %s", err),
		}
	}
	log.Debugf("params for infra:\n%s", paramsMainBytes)

	output, err := c.infrastructureProvider.RunCommand("apply", &paramsMainBytes)
	if err != nil {
		return []error{
			fmt.Errorf("Error while running infrastructure provider: %s", err),
		}
	}

	// check and merge output from infrastructure apply
	err = c.Parameters.Parse(string(output))
	if err != nil {
		return []error{
			fmt.Errorf("Error while parsing infrastructure provider output: %s", err),
		}
	}

	// store inventory
	c.WriteInventory()
	if err != nil {
		return []error{
			fmt.Errorf("Error while storing inventory: %s", err),
		}
	}

	errs = append(errs, c.Parameters.Validate()...)
	if len(errs) > 0 {
		return errs
	}

	// run config apply
	paramsMainBytes, err = yaml.Marshal(c.Parameters)
	if err != nil {
		return []error{
			fmt.Errorf("Error while writing parameters file: %s", err),
		}
	}
	log.Debugf("params after merge with output from infra:\n%s", paramsMainBytes)

	output, err = c.configProvider.RunCommand("apply", &paramsMainBytes)
	if err != nil {
		return []error{
			fmt.Errorf("Error while running infratstructure provider: %s", err),
		}
	}

	return []error{}
}
