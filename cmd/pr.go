package cmd

import (
	"context"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"os/exec"
)

var prCmd = &cobra.Command{
	Use:   "update",
	RunE:  update,
	Short: "clone, dependencies update and pull request on github",
}

func init() {
	RootCmd.AddCommand(prCmd)
}

func update(cmd *cobra.Command, _ []string) error {
	tmp, err := ioutil.TempDir("", "depbot")
	if err != nil {
		return errors.Wrapf(err, "cannot create temporary directory with prefix '%s'", "depbot")
	}
	defer os.Remove(tmp)
	ctx, err := createContext(cmd, tmp)
	if err != nil {
		return err
	}
	mkWoringDir(ctx.pwd)

	err = ctx.scm.clone()
	if err != nil {
		return errors.Wrap(err, "cannot clone")
	}

	err = ctx.updater.update()
	if err != nil {
		return errors.Wrap(err, "cannot update")
	}

	err = ctx.scm.checkout()
	if err != nil {
		return errors.Wrap(err, "cannot checkout")
	}

	err = ctx.scm.commit()
	if err != nil {
		return errors.Wrap(err, "cannot commit")
	}

	err = ctx.hub.fork()
	if err != nil {
		return errors.Wrap(err, "cannot fork")
	}

	err = ctx.scm.push()
	if err != nil {
		return errors.Wrap(err, "cannot push")
	}

	err = ctx.hub.pullRequest()
	if err != nil {
		return errors.Wrap(err, "cannot create pull request")
	}
	return nil
}

func mkWoringDir(workingDir string) {
	mkdir := exec.CommandContext(context.Background(), "mkdir", "-p", workingDir)
	mkdir.Stdout = os.Stdout
	mkdir.Stderr = os.Stderr
	mkdir.Run()
	logrus.Infof("Working directory is '%s'", workingDir)
}
