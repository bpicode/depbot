package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/bpicode/depbot/cmd"
)

func main() {
	err := cmd.RootCmd.Execute()
	if err != nil {
		logrus.Fatal(err)
	}
}
