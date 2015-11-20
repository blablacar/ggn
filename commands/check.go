package commands

import (
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
)

func checkEnv(cmd *cobra.Command, args []string, work *work.Work, envString string) {
	work.LoadEnv(envString).Check()
}

func checkService(cmd *cobra.Command, args []string, work *work.Work, env string, serviceName string) {
	work.LoadEnv(env).LoadService(serviceName).Check()
}
