package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/work"
	"github.com/spf13/cobra"
)

func loadEnvCommandsReturnNewRoot(osArgs []string, rootCmd *cobra.Command) *cobra.Command {
	logrus.WithField("path", ggn.Home.Config.WorkPath).Debug("Loading envs")
	work := work.NewWork(ggn.Home.Config.WorkPath)

	newRootCmd := &cobra.Command{
		Use: "ggn",
	}

	for _, envNames := range work.ListEnvs() {
		env := work.LoadEnv(envNames)

		envCmd := &cobra.Command{
			Use:   env.GetName(),
			Short: "Run command for " + env.GetName(),
			Run: func(cmd *cobra.Command, args []string) {

				newRootCmd.AddCommand(prepareEnvCommands(env))
				newRootCmd.SetArgs(osArgs[1:])
				newRootCmd.Execute()
			},
		}
		rootCmd.AddCommand(envCmd)
	}
	return newRootCmd
}
