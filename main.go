package main

import (
	"os"

	"github.com/simonswine/slingshot/pkg/slingshot"
)

func main() {
	slingshot := slingshot.NewSlingshot()
	slingshot.App.Run(os.Args)
}
