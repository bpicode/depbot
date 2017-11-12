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
	RootCmd.InitDefaultHelpFlag()
	RootCmd.InitDefaultHelpCmd()
}
