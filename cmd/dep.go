package cmd

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"os"
	"os/exec"
)

type dep struct {
	pwd    string
	gopath string
}

func (d *dep) update() error {
	logrus.Infof("Updating dependencies")
	err := d.command("dep", "ensure", "-update").Run()
	return errors.Wrap(err, "dep ensure -update failed")
}

func (d *dep) command(name string, arg ...string) *exec.Cmd {
	c := exec.CommandContext(context.Background(), name, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = d.pwd
	c.Env = append(os.Environ(), "GOPATH="+d.gopath)
	return c
}
