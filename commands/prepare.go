package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/work"
	"github.com/spf13/cobra"
)

func loadEnvCommands(rootCmd *cobra.Command) {
	logrus.WithField("path", ggn.Home.Config.WorkPath).Debug("Loading envs")
	work := work.NewWork(ggn.Home.Config.WorkPath)

	for _, envNames := range work.ListEnvs() {
		env := work.LoadEnv(envNames)
		rootCmd.AddCommand(prepareEnvCommands(env))
	}

}
