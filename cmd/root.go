package cmd

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:           "depbot [subcommand]",
	SilenceUsage:  true,
	SilenceErrors: true,
}

func init() {
	cobra.OnInitialize()
	RootCmd.PersistentFlags().String("project", "", "project, e.g. github.com/bpicode/depbot")
	RootCmd.PersistentFlags().String("token", "", "github token")
	RootCmd.PersistentFlags().String("user", "dep-bot", "github user name")
	RootCmd.PersistentFlags().String("mail", "depbot@yandex.com", "github user email")
	RootCmd.InitDefaultHelpFlag()
	RootCmd.InitDefaultHelpCmd()
}
