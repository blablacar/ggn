package commands

import (
	"github.com/blablacar/green-garden/config"
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
)

var generateCmd = &cobra.Command{
	Use:   "generate",
	Short: "generate units for all envs",
	Run: func(cmd *cobra.Command, args []string) {
		work := work.NewWork(config.GetConfig().WorkPath)
		for _, envName := range work.ListEnvs() {
			env := work.LoadEnv(envName)
			env.Generate()
		}
	},
}

func generateService(cmd *cobra.Command, args []string, work *work.Work, env string, service string) {
	work.LoadEnv(env).LoadService(service).Generate(args)
}

func generateEnv(cmd *cobra.Command, args []string, work *work.Work, env string) {
	work.LoadEnv(env).Generate()
}
