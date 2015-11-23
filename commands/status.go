package commands

import (
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
)

func statusEnv(cmd *cobra.Command, args []string, work *work.Work, env string) {
	work.LoadEnv(env).Status()
}

func statusService(cmd *cobra.Command, args []string, work *work.Work, env string, service string) {
	work.LoadEnv(env).LoadService(service).Status()
}
