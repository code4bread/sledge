package main

import (
	"github.com/code4bread/sledge/cmd"
	"github.com/sirupsen/logrus"
)

func main() {
	log := logrus.New()
	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
