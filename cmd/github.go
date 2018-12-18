package cmd

import (
	"context"
	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/google/go-github/github"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
	"os"
	"os/exec"
)

type ghub struct {
	user    string
	token   string
	pwd     string
	repo    string
	owner   string
	fbranch string
}

func (g *ghub) fork() error {
	client := g.client()
	_, err := client.Repositories.Delete(context.Background(), g.user, g.repo)
	if err != nil {
		logrus.Infof("Could not delete remote fork: %s", err)
	}

	_, _, err = client.Repositories.CreateFork(context.Background(), g.owner, g.repo, &github.RepositoryCreateForkOptions{})
	if err != nil {
		logrus.Infof("Could not fork (sometimes you may ignore this error): %s", err)
	}

	err = g.command("git", "remote", "add", "fork",
		fmt.Sprintf("https://%s:%s@github.com/%s/%s", g.user, g.token, g.user, g.repo)).Run()
	if err != nil {
		return errors.Wrap(err, "cannot add fork as remote")
	}
	return nil
}

func (g *ghub) client() *github.Client {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: g.token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	client := github.NewClient(tc)
	return client
}

func (g *ghub) pullRequest() error {
	client := g.client()

	title := "Automatic upgrade of dependencies by depbot"
	head := fmt.Sprintf("%s:%s", g.user, g.fbranch)
	base := "master"
	body := "Automatic upgrade of dependencies. Please review carefully. I will not answer to review comments, because I am a bot."

	_, _, err := client.PullRequests.Create(context.Background(), g.owner, g.repo, &github.NewPullRequest{
		Title: &title,
		Head:  &head,
		Base:  &base,
		Body:  &body,
	})
	return errors.Wrap(err, "could not create pull request")
}

func (g *ghub) command(name string, arg ...string) *exec.Cmd {
	c := exec.CommandContext(context.Background(), name, arg...)
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	c.Dir = g.pwd
	c.Env = os.Environ()
	return c
}
