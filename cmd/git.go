package cmd

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"os"
	"os/exec"
)

type git struct {
	pwd     string
	project string
	user    string
	email   string
	auth    string
	fbranch string
}

func (g *git) clone() error {
	logrus.Infof("Cloning '%s'", g.project)
	c := g.command("git", "clone", "--depth=1",
		fmt.Sprintf("https://%s:%s@%s", g.user, g.auth, g.project), ".")
	return c.Run()
}

func (g *git) checkout() error {
	logrus.Infof("Switch to branch '%s'", g.fbranch)
	err := g.command("git", "checkout", "-b", g.fbranch).Run()
	if err != nil {
		logrus.Warnf("Checkout of new branch failed: %v", err)
	}
	err = g.command("git", "checkout", g.fbranch).Run()
	return errors.Wrapf(err, "could not switch to a branch '%s'", g.fbranch)
}

func (g *git) commit() error {
	logrus.Infof("Committing changes")
	g.command("git", "config", "user.email", g.email).Run()
	g.command("git", "config", "user.name", g.user).Run()
	err := g.command("git", "add", "--all").Run()
	if err != nil {
		return errors.Wrap(err, "git add --all failed")
	}

	err = g.command("git", "commit", "--message=Upgrade of 3rd party dependencies").Run()
	return errors.Wrap(err, "git commit failed")
}

func (g *git) push() error {
	err := g.command("git", "push", "fork").Run()
	return errors.Wrap(err, "could not push")
}

func (g *git) command(name string, arg ...string) *exec.Cmd {
	c := exec.CommandContext(context.Background(), name, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = g.pwd
	c.Env = os.Environ()
	return c
}
