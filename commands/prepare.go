package commands

import (
	"github.com/Sirupsen/logrus"
	"github.com/blablacar/ggn/ggn"
	"github.com/blablacar/ggn/work"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate units for all envs",
	Run: func(cmd *cobra.Command, args []string) {
		work := work.NewWork(ggn.Home.Config.WorkPath)
		for _, envName := range work.ListEnvs() {
			env := work.LoadEnv(envName)
			env.Generate()
		}
	},
}

func loadEnvCommands(rootCmd *cobra.Command) {
	logrus.WithField("path", ggn.Home.Config.WorkPath).Debug("Loading envs")
	work := work.NewWork(ggn.Home.Config.WorkPath)

	for _, envNames := range work.ListEnvs() {
		env := work.LoadEnv(envNames)
		rootCmd.AddCommand(prepareEnvCommands(env))
	}

}
