package cmd

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"fmt"
	"io/ioutil"

	"github.com/google/go-github/github"
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"golang.org/x/oauth2"
)

var prCmd = &cobra.Command{
	Use:  "pr",
	RunE: pr,
}

func init() {
	RootCmd.AddCommand(prCmd)
}

func pr(cmd *cobra.Command, _ []string) error {
	tmp, err := ioutil.TempDir("", "depbot")
	if err != nil {
		return errors.Wrapf(err, "cannot create temporary directory with prefix '%s'", "depbot")
	}
	defer os.Remove(tmp)
	ctx, err := createContext(cmd, tmp)
	if err != nil {
		return err
	}

	ctx.pwd = ctx.gopath + "/src/" + ctx.project
	mkWoringDir(ctx.pwd)

	err = clone(ctx)
	if err != nil {
		return errors.Wrapf(err, "cannot clone '%s'", ctx.project)
	}

	err = depUpdate(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot update")
	}

	err = checkoutBranch(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot checkout")
	}

	err = gitCommit(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot commit")
	}

	err = fork(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot fork")
	}

	err = gitPush(ctx)
	if err != nil {
		return errors.Wrap(err, "cannot push")
	}

	err = pullRequest(ctx)
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

func pullRequest(ctx *projectCtx) error {
	client := ctx.githubClient()

	title := "Automatic upgrade of dependencies by depbot"
	head := fmt.Sprintf("%s:%s", ctx.user, ctx.featurebranch)
	base := "master"
	body := "Bot upgrade of dependencies using dep. Please review carefully. I will not answer to review comments."

	_, _, err := client.PullRequests.Create(context.Background(), ctx.owner, ctx.repo, &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	})
	return errors.Wrap(err, "could not create pull request")
}
func gitPush(ctx *projectCtx) error {
	err := ctx.command("git", "push", "fork").Run()
	return errors.Wrap(err, "could not push")
}

func fork(ctx *projectCtx) error {
	client := ctx.githubClient()
	_, err := client.Repositories.Delete(context.Background(), ctx.user, ctx.repo)
	if err != nil {
		logrus.Infof("Could not delete remote fork: %s", err)
	}

	_, _, err = client.Repositories.CreateFork(context.Background(), ctx.owner, ctx.repo, &github.RepositoryCreateForkOptions{})
	if err != nil {
		logrus.Infof("Could not fork (sometimes you may ignore this error): %s", err)
	}

	err = ctx.command("git", "remote", "add", "fork",
		fmt.Sprintf("https://%s:%s@github.com/%s/%s", ctx.user, ctx.token, ctx.user, ctx.repo)).Run()
	if err != nil {
		return errors.Wrap(err, "cannot add fork as remote")
	}
	return nil
}

func gitCommit(ctx *projectCtx) error {
	logrus.Infof("Committing changes")
	ctx.command("git", "config", "user.email", ctx.email).Run()
	ctx.command("git", "config", "user.name", ctx.user).Run()
	err := ctx.command("git", "add", "--all").Run()
	if err != nil {
		return errors.Wrap(err, "git add --all failed")
	}

	err = ctx.command("git", "commit", "--message=automatic upgrade of 3rd party dependencies").Run()
	return errors.Wrap(err, "git commit failed")
}

func depUpdate(ctx *projectCtx) error {
	logrus.Infof("Updating dependencies")
	err := ctx.command("dep", "ensure", "-update").Run()
	if err != nil {
		return errors.Wrap(err, "dep ensure -update failed")
	}
	err = ctx.command("dep", "prune").Run()
	return errors.Wrap(err, "dep prune failed")
}

func checkoutBranch(ctx *projectCtx) error {
	logrus.Infof("Switch to branch '%s'", ctx.featurebranch)
	err := ctx.command("git", "checkout", "-b", ctx.featurebranch).Run()
	if err != nil {
		logrus.Warnf("Checkout of new branch failed: %v", err)
	}
	err = ctx.command("git", "checkout", ctx.featurebranch).Run()
	return errors.Wrapf(err, "could not switch to a branch '%s'", ctx.featurebranch)
}

func clone(ctx *projectCtx) error {
	logrus.Infof("Cloning '%s'", ctx.project)
	c := ctx.command("git", "clone", "--depth=1",
		fmt.Sprintf("https://%s:%s@%s", ctx.user, ctx.token, ctx.project), ".")
	return c.Run()
}

type projectCtx struct {
	project       string
	token         string
	user          string
	email         string
	pwd           string
	repo          string
	owner         string
	featurebranch string
	gopath        string
}

func createContext(cmd *cobra.Command, gopath string) (*projectCtx, error) {
	ctx := new(projectCtx)
	ctx.gopath = gopath

	ctx.user = cmd.Flag("user").Value.String()
	ctx.email = cmd.Flag("mail").Value.String()

	project := cmd.Flag("project").Value.String()
	if project == "" {
		return nil, errors.New("required flag 'repo' not set")
	}
	ctx.project = project
	token := cmd.Flag("token").Value.String()
	if token == "" {
		return nil, errors.New("required flag 'token' not set")
	}
	ctx.token = token

	projectPaths := strings.Split(ctx.project, "/")
	ctx.repo = projectPaths[len(projectPaths)-1]
	ctx.owner = projectPaths[1]

	ctx.featurebranch = "auto-upgrade-dependencies"

	return ctx, nil
}

func (ctx *projectCtx) command(name string, arg ...string) *exec.Cmd {
	c := exec.CommandContext(context.Background(), name, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = ctx.pwd
	c.Env = append(os.Environ(), "GOPATH="+ctx.gopath)
	return c
}

func (ctx *projectCtx) githubClient() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: ctx.token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return client
}
