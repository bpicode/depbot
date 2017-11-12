package cmd

import (
	"context"
	"os"
	"os/exec"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
)

var prCmd = &cobra.Command{
	Use:  "pr",
	RunE: pr,
}

var gopath string

var branch = "auto-upgrade-dependencies"

var workingDir string

func init() {
	RootCmd.AddCommand(prCmd)
}

func pr(cmd *cobra.Command, _ []string) error {
	project := cmd.Flag("project").Value.String()
	if project == "" {
		return errors.New("required flag 'repo' not set")
	}

	tmp, err := ioutil.TempDir("", "depbot")
	if err != nil {
		return errors.Wrapf(err, "cannot create temporary directory with prefix '%s'", "depbot")
	}
	defer os.Remove(tmp)
	gopath = tmp

	command("mkdir", gopath).Run()
	workingDir = gopath + "/src/" + project
	logrus.Infof("Working directory is '%s'", workingDir)

	_ = checkoutMaster()

	err = goGet(project)
	if err != nil {
		return errors.Wrapf(err, "cannot get project '%s'", project)
	}

	err = checkoutMaster()
	if err != nil {
		logrus.Warn("Reset to master branch failed", err)
	}

	err = depUpdate()
	if err != nil {
		return errors.Wrapf(err, "cannot update dependencies of project '%s'", project)
	}

	// git push
	// create PR
	err = checkoutBranch()
	if err != nil {
		return errors.Wrap(err, "cannot checkout branch")
	}

	err = gitCommit()
	if err != nil {
		return errors.Wrap(err, "cannot commit changes")
	}

	err = gitPush()
	return err
}

func gitPush() error {
	logrus.Infof("Deleting upstream")
	command("git", "push", "origin", "--delete", "auto-upgrade-dependencies").Run()
	logrus.Infof("Pushing changes")
	err := command("git", "push", "--set-upstream", "origin", branch).Run()
	return errors.Wrap(err, "failed to push to remote repository")
}

func gitCommit() error {
	logrus.Infof("Committing changes")
	command("git", "config", "user.email", "depbot@yandex.com").Run()
	command("git", "config", "user.name", "dep-bot").Run()
	err := command("git", "add", "--all").Run()
	if err != nil {
		return errors.Wrap(err, "git add --all failed")
	}

	err = command("git", "commit", "--message=automatic upgrade of 3rd party dependencies").Run()
	return errors.Wrap(err, "git commit failed")
}

func depUpdate() error {
	logrus.Infof("Updating dependencies")
	err := command("dep", "ensure", "-update").Run()
	return errors.Wrap(err, "dep ensure -update failed")
}

func checkoutBranch() error {
	logrus.Infof("Switch to branch '%s'", branch)
	err := command("git", "checkout", "-b", branch).Run()
	if err != nil {
		logrus.Warnf("Checkout of new branch failed: %v", err)
	}
	err = command("git", "checkout", branch).Run()
	return errors.Wrapf(err, "could not switch to a branch '%s'", branch)
}

func goGet(project string) error {
	logrus.Infof("Go-getting '%s'", project)
	c := command("go", "get", "-u", project)
	c.Dir = gopath
	return c.Run()
}

func checkoutMaster() error {
	logrus.Infof("Switching to master branch")
	err := command("git", "checkout", "master").Run()
	if err != nil {
		return errors.Wrap(err, "cannot checkout master branch")
	}
	command("git", "branch", "-D", "auto-upgrade-dependencies").Run()
	return nil
}

func command(name string, arg ...string) *exec.Cmd {
	c := exec.CommandContext(context.Background(), name, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = workingDir
	c.Env = append(os.Environ(), "GOPATH="+gopath)
	return c
}
