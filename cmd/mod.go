package cmd

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"os"
	"os/exec"
)

type mod struct {
	pwd       string
	gopath    string
	patchOnly bool
}

func (m *mod) update() error {
	logrus.Infof("Updating dependencies")
	if m.patchOnly {
		err := m.command("go", "get", "-u=patch").Run()
		if err != nil {
			return errors.Wrap(err, "go get -u=patch failed")
		}
	} else {
		err := m.command("go", "get", "-u").Run()
		if err != nil {
			return errors.Wrap(err, "go get -u failed")
		}
	}
	err := m.command("go", "mod", "tidy").Run()
	if err != nil {
		return errors.Wrap(err, "go mod tidy failed")
	}
	err = m.command("go", "mod", "vendor").Run()
	return errors.Wrap(err, "go mod vendor failed")
}

func (m *mod) command(name string, arg ...string) *exec.Cmd {
	c := exec.CommandContext(context.Background(), name, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = m.pwd
	c.Env = append(os.Environ(), "GOPATH="+m.gopath, "GO111MODULE=on")
	return c
}
