package cmd

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"strings"
)

type projectCtx struct {
	scm     scm
	hub     hub
	updater updater
	pwd     string
}

type scm interface {
	clone() error
	checkout() error
	commit() error
	push() error
}

type hub interface {
	fork() error
	pullRequest() error
}

type updater interface {
	update() error
}

func createContext(cmd *cobra.Command, gopath string) (*projectCtx, error) {
	user := cmd.Flag("user").Value.String()
	mail := cmd.Flag("mail").Value.String()
	project := cmd.Flag("project").Value.String()
	if project == "" {
		return nil, errors.New("required flag 'repo' not set")
	}
	token := cmd.Flag("token").Value.String()
	if token == "" {
		return nil, errors.New("required flag 'token' not set")
	}
	branch := "auto-upgrade-dependencies"
	pwd := gopath + "/src/" + project

	projectPaths := strings.Split(project, "/")
	repo := projectPaths[len(projectPaths)-1]
	owner := projectPaths[1]

	var updater updater
	mode := cmd.Flag("mode").Value.String()
	if mode == "gomodule" {
		updater = &mod{
			pwd:       pwd,
			gopath:    gopath,
			patchOnly: cmd.Flag("raise").Value.String() == "patch",
		}
	} else {
		updater = &dep{
			pwd:    pwd,
			gopath: gopath,
		}
	}

	ctx := &projectCtx{
		scm: &git{
			pwd:     pwd,
			project: project,
			user:    user,
			email:   mail,
			auth:    token,
			fbranch: branch,
		},
		hub: &ghub{
			pwd:     pwd,
			repo:    repo,
			user:    user,
			token:   token,
			fbranch: branch,
			owner:   owner,
		},
		updater: updater,
		pwd:     pwd,
	}

	return ctx, nil
}
