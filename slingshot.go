package main

import (
	"io/ioutil"
	"path/filepath"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/fsouza/go-dockerclient"
	"github.com/simonswine/slingshot/utils"
	"os"
)

const SlingshotClusterFileName = "cluster.yaml"

type Slingshot struct {
	dockerClient *docker.Client
	clusters     []*Cluster
	configDir    string
}

func NewSlingshot() *Slingshot {
	s := &Slingshot{}
	return s
}

func (s *Slingshot) Init() {
	s.log().Infof("initialise %s %s (%s)", AppName, AppVersion, GitCommit)
	s.ensureConfigDir()
	s.loadClusters()
}

func (s *Slingshot) loadClusters() {
	files, _ := ioutil.ReadDir(s.configDir)
	for _, f := range files {
		if f.IsDir() {
			configPath := filepath.Join(
				s.configDir,
				f.Name(),
				SlingshotClusterFileName,
			)

			// skip if a dir or not exists
			stat, err := os.Stat(configPath)
			if err != nil || stat.IsDir() {
				continue
			}

			// load cluster otherwise
			c, err := LoadClusterFromPath(s, configPath)
			if err != nil {
				s.log().Warnf("Could not read cluster in '%s': %s", configPath, err)
			}

			s.clusters = append(s.clusters, c)
			s.log().Debugf("read cluster config file in '%s'", configPath)
		}
	}
}

func (s *Slingshot) ensureConfigDir() {
	configDir, err := utils.SlinshotConfigDirPath()
	if err != nil {
		s.log().Fatal("failed to detect home directory: ", err)
	}
	s.configDir = configDir

	if err := utils.EnsureDirectory(s.configDir); err != nil {
		s.log().Fatal(err)
	}
}

func (s *Slingshot) log() *log.Entry {
	return log.WithFields(log.Fields{
		"context": "slingshot",
	})
}

func (s *Slingshot) Docker() (*docker.Client, error) {
	if s.dockerClient != nil {
		return s.dockerClient, nil
	}
	client, err := docker.NewClientFromEnv()
	if err != nil {
		return nil, err
	}

	env, err := client.Version()
	if err != nil {
		s.log().Fatal("connecting to docker failed: ", err)
		return nil, err
	}
	s.log().Debugf("connected to docker %+v", env.Get("Version"))

	s.dockerClient = client
	return s.dockerClient, nil
}

func (s *Slingshot) clusterCreateAction(context *cli.Context) {
	c := NewCluster(s)

	s.clusters = append(s.clusters, c)

	errs := c.Create(context)
	if len(errs) > 0 {
		for _, err := range errs {
			log.Error(err)
		}
		log.Fatal("errors prevent further execution")
	}
}

func (s *Slingshot) unimplementedAction(context *cli.Context) {
	s.log().Warnf("command '%s' (%s) not implemented", context.Command.HelpName, context.Command.Usage)
}

func (s *Slingshot) clusterCommands() []cli.Command {
	return []cli.Command{
		{
			Name:   "create",
			Usage:  "add a new cluster",
			Action: s.clusterCreateAction,
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "infrastructure-provider, I",
					Usage: "Image name of the infrastructure provider to use",
				},
				cli.StringFlag{
					Name:  "config-provider, C",
					Usage: "Image name of the config provider to use",
				},
				cli.StringFlag{
					Name:  "ssh-key, i",
					Usage: "SSH private key to use (please provide an uncrypted key, default: vagrant insecure key)",
				},
			},
		},
		{
			Name:  "list",
			Usage: "list existing clusters",
			// TODO Implement me
			Action: s.unimplementedAction,
		},
	}
}

func (s *Slingshot) Commands() []cli.Command {
	return []cli.Command{
		{
			Name:        "cluster",
			Usage:       "manage clusters",
			Subcommands: s.clusterCommands(),
			Before: func(context *cli.Context) error {
				s.Init()
				return nil
			},
		},
	}
}
