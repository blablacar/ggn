package commands

import (
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
	"os"
)

func update(cmd *cobra.Command, args []string, work *work.Work, env string, service string) {
	err := work.LoadEnv(env).LoadService(service).Update()
	if err != nil {
		os.Exit(1)
	}
}
