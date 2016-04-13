package main

import (
	"os"

	log "github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
)

func main() {
	log.SetLevel(log.DebugLevel)

	slingshot := NewSlingshot()

	app := cli.NewApp()
	app.Name = AppName
	app.Version = AppVersion
	app.Usage = "yet another zero to kubernetes utility"
	app.Commands = slingshot.Commands()

	app.Run(os.Args)
}
