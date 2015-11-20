package commands

import (
	"github.com/blablacar/green-garden/work"
	"github.com/spf13/cobra"
)

func fleetctl(cmd *cobra.Command, args []string, work *work.Work, env string) {
	work.LoadEnv(env).Fleetctl(args)
}
